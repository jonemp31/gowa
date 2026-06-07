package helpers

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	"github.com/sirupsen/logrus"
)

// reconnectCooldown guards against reconnect storms: a device that was already
// attempted recently is skipped until the cooldown expires.
// The cooldown duration is read from config.ReconnectCooldownSeconds
// (env RECONNECT_COOLDOWN_SECONDS, default 120) so it can be tuned without
// recompiling. Reading at call time (not package init) ensures env overrides
// applied by viper in cobra.OnInitialize are always respected.
var (
	reconnectCooldownMu  sync.Mutex
	reconnectCooldownMap = make(map[string]time.Time)
)

func reconnectCooldownDuration() time.Duration {
	secs := config.ReconnectCooldownSeconds
	if secs <= 0 {
		secs = 120
	}
	return time.Duration(secs) * time.Second
}

func isReconnectOnCooldown(deviceID string) bool {
	reconnectCooldownMu.Lock()
	defer reconnectCooldownMu.Unlock()
	last, ok := reconnectCooldownMap[deviceID]
	return ok && time.Since(last) < reconnectCooldownDuration()
}

func markReconnectAttempt(deviceID string) {
	reconnectCooldownMu.Lock()
	reconnectCooldownMap[deviceID] = time.Now()
	reconnectCooldownMu.Unlock()
}

func SetAutoConnectAfterBooting(service domainApp.IAppUsecase) {
	time.Sleep(2 * time.Second)
	devices, err := service.FetchDevices(context.Background())
	if err != nil || len(devices) == 0 {
		logrus.Warn("auto-connect skipped: no devices available")
		return
	}

	maxWorkersVal := config.BootMaxWorkers
	if maxWorkersVal <= 0 {
		maxWorkersVal = 5
	}
	delayMs := config.BootDeviceDelayMs
	if delayMs <= 0 {
		delayMs = 1200
	}
	launchDelay := time.Duration(delayMs) * time.Millisecond
	total := len(devices)

	logrus.Infof("[AUTO-CONNECT] starting boot sequence: %d device(s), workers=%d, delay=%s",
		total, maxWorkersVal, launchDelay)

	sem := make(chan struct{}, maxWorkersVal)
	var wg sync.WaitGroup
	var completed atomic.Int32

	for i, device := range devices {
		if i > 0 {
			time.Sleep(launchDelay)
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(d domainApp.DevicesResponse) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := service.Reconnect(context.Background(), d.Device); err != nil {
				// Device was never logged in (no JID) and old enough — purge on boot.
				// Grace period of 30 min protects devices that were just created and are
				// waiting for a QR scan before the API was restarted.
				if strings.Contains(err.Error(), "session deleted") && d.JID == "" {
					age := time.Since(d.CreatedAt)
					if age >= 30*time.Minute {
						logrus.Infof("[AUTO-CONNECT] device %s never logged in (%s old) — purging on boot", d.Device, age.Round(time.Minute))
						if purgeErr := service.Logout(context.Background(), d.Device); purgeErr != nil {
							logrus.Warnf("[AUTO-CONNECT] failed to purge device %s: %v", d.Device, purgeErr)
						}
					} else {
						logrus.Infof("[AUTO-CONNECT] device %s never logged in but only %s old — skipping purge", d.Device, age.Round(time.Minute))
					}
				} else {
					logrus.Warnf("[AUTO-CONNECT] failed for device %s: %v", d.Device, err)
				}
			} else {
				logrus.Infof("[AUTO-CONNECT] connected device %s", d.Device)
			}

			// Progress counter — logged after every device regardless of outcome.
			n := int(completed.Add(1))
			pct := n * 100 / total
			logrus.Infof("[AUTO-CONNECT] %d/%d devices completed (%d%%)", n, total, pct)
		}(device)
	}

	// Do not block the caller. The API starts accepting requests as soon as Fiber
	// binds — boot connections happen in the background. The wg.Wait goroutine
	// only exists to emit a completion log once every device is processed.
	go func() {
		wg.Wait()
		logrus.Infof("[AUTO-CONNECT] boot sequence finished — all %d device(s) processed", total)
	}()
}

func SetAutoReconnectChecking(service domainApp.IAppUsecase) {
	go func() {
		// Ticker ensures a fixed 60s interval between starts regardless of how long
		// each iteration takes — unlike time.Sleep which adds processing time on top.
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Prune cooldown entries older than 2× the cooldown window so device IDs
			// from deleted or purged devices don't accumulate indefinitely in memory.
			reconnectCooldownMu.Lock()
			for id, ts := range reconnectCooldownMap {
				if time.Since(ts) > 2*reconnectCooldownDuration() {
					delete(reconnectCooldownMap, id)
				}
			}
			reconnectCooldownMu.Unlock()

			devices, err := service.FetchDevices(context.Background())
			if err != nil || len(devices) == 0 {
				continue
			}

			// Phase 1: check all devices serially (fast — in-memory reads only).
			// Collecting first avoids blocking Status() calls behind reconnect timeouts.
			var toReconnect []domainApp.DevicesResponse
			for _, device := range devices {
				isConnected, _, _ := service.Status(context.Background(), device.Device)
				if isConnected {
					continue
				}
				if isReconnectOnCooldown(device.Device) {
					logrus.Debugf("[AUTO-RECONNECT] device %s on cooldown, skipping", device.Device)
					continue
				}
				markReconnectAttempt(device.Device)
				toReconnect = append(toReconnect, device)
			}

			// Phase 2: reconnect in parallel, bounded by config.ReconnectMaxWorkers
			// (env RECONNECT_MAX_WORKERS, default 10).
			reconnectWorkers := config.ReconnectMaxWorkers
			if reconnectWorkers <= 0 {
				reconnectWorkers = 10
			}
			sem := make(chan struct{}, reconnectWorkers)
			var wg sync.WaitGroup

			for _, device := range toReconnect {
				sem <- struct{}{}
				wg.Add(1)

				go func(dev domainApp.DevicesResponse) {
					defer wg.Done()
					defer func() { <-sem }()

					if err := service.Reconnect(context.Background(), dev.Device); err != nil {
						if strings.Contains(err.Error(), "session deleted") {
							age := time.Since(dev.CreatedAt)
							if dev.JID == "" {
								// No JID means this device was never logged in (no QR scan completed).
								// Use a short grace period so a freshly created device waiting for
								// QR scan is not immediately purged if the checker fires first.
								if age < 30*time.Minute {
									logrus.Infof("[AUTO-RECONNECT] device %s never logged in, created %s ago — skipping purge (grace period)",
										dev.Device, age.Round(time.Minute))
								} else {
									logrus.Infof("[AUTO-RECONNECT] device %s never logged in, created %s ago — purging",
										dev.Device, age.Round(time.Minute))
									if purgeErr := service.Logout(context.Background(), dev.Device); purgeErr != nil {
										logrus.Warnf("[AUTO-RECONNECT] failed to purge device %s: %v", dev.Device, purgeErr)
									}
								}
							} else {
								// Device had a session (JID known) but it was revoked from the phone.
								// Keep the 2h protection before purging to avoid false positives from
								// transient server-side errors.
								if age < 2*time.Hour {
									logrus.Infof("[AUTO-RECONNECT] session deleted for device %s but created %s ago — skipping auto-purge",
										dev.Device, age.Round(time.Minute))
								} else {
									logrus.Infof("[AUTO-RECONNECT] session deleted for device %s (created %s ago) — removing",
										dev.Device, age.Round(time.Minute))
									if purgeErr := service.Logout(context.Background(), dev.Device); purgeErr != nil {
										logrus.Warnf("[AUTO-RECONNECT] failed to remove session-deleted device %s: %v", dev.Device, purgeErr)
									}
								}
							}
						} else {
							logrus.Warnf("[AUTO-RECONNECT] failed for device %s: %v", dev.Device, err)
						}
					} else {
						logrus.Infof("[AUTO-RECONNECT] recovered device %s", dev.Device)
					}
				}(device)
			}

			wg.Wait()
		}
	}()
}

// presenceWindow defines a time window during which a device may appear online.
// Each device draws its own random firing time within [startH:startM, endH:endM]
// and a random online duration within [minSec, maxSec]. skipPct is the percent
// chance (0–100) that a device skips this window entirely for that day,
// simulating a busy or distracted user.
type presenceWindow struct {
	name    string
	startH  int
	startM  int
	endH    int
	endM    int
	minSec  int
	maxSec  int
	skipPct int
}

// dailyPresenceWindows defines 8 human-like activity slots spread across the day.
// Silence is enforced between 23:30 and 07:00 (sleeping hours).
// Window durations and skip rates are tuned to produce ~6–8 signals/day
// and ~10–30 min total online time per device.
var dailyPresenceWindows = []presenceWindow{
	{"wake_up", 7, 0, 8, 30, 15, 45, 20},    // quick check on waking
	{"morning_commute", 8, 45, 9, 30, 20, 60, 20},  // coffee / commute scroll
	{"morning_break", 10, 30, 11, 30, 30, 90, 20},  // mid-morning work break
	{"lunch", 12, 0, 13, 30, 120, 300, 15},          // lunch — longer session
	{"afternoon_break", 15, 0, 16, 0, 20, 60, 20},  // afternoon pause
	{"end_of_work", 17, 30, 19, 0, 60, 180, 20},    // leaving work / commute
	{"evening", 20, 0, 22, 0, 180, 480, 10},         // main evening session (low skip)
	{"pre_sleep", 22, 30, 23, 30, 30, 90, 20},       // last check before bed
}

// SetDailyPresenceScheduler runs a background goroutine that simulates human-like
// WhatsApp presence for every registered device. Each device independently draws
// random firing times and online durations across 8 daily windows (07:00–23:30),
// with a per-window skip probability that mimics busy or distracted days.
// The scheduler recalculates at each local midnight to stay accurate across DST changes.
func SetDailyPresenceScheduler(service domainApp.IAppUsecase) {
	go func() {
		for {
			now := time.Now()
			scheduleDayPresences(service, now)
			tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, now.Location())
			time.Sleep(time.Until(tomorrow))
		}
	}()
}

func scheduleDayPresences(service domainApp.IAppUsecase, now time.Time) {
	devices, err := service.FetchDevices(context.Background())
	if err != nil || len(devices) == 0 {
		return
	}
	for _, device := range devices {
		d := device
		go scheduleDevicePresences(service, d, now)
	}
}

func scheduleDevicePresences(service domainApp.IAppUsecase, device domainApp.DevicesResponse, now time.Time) {
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	for _, w := range dailyPresenceWindows {
		// Each window is evaluated independently per device.
		// Capture loop variable for the goroutine closure.
		win := w

		// ~skipPct% chance to skip this window (simulates a busy day).
		if rand.Intn(100) < win.skipPct {
			continue
		}

		// Draw a random fire time within [windowStart, windowEnd].
		windowStart := today.Add(time.Duration(win.startH)*time.Hour + time.Duration(win.startM)*time.Minute)
		windowEnd := today.Add(time.Duration(win.endH)*time.Hour + time.Duration(win.endM)*time.Minute)
		span := windowEnd.Sub(windowStart)
		fireAt := windowStart.Add(time.Duration(rand.Int63n(int64(span))))

		// Skip windows that have already passed today (e.g. scheduler started mid-day).
		if !fireAt.After(now) {
			continue
		}

		// Draw random online duration within [minSec, maxSec].
		onlineSec := win.minSec + rand.Intn(win.maxSec-win.minSec+1)

		go func(at time.Time, seconds int, label string) {
			time.Sleep(time.Until(at))
			ctx := context.Background()
			if err := service.SendDevicePresence(ctx, device.Device, "available"); err != nil {
				logrus.Warnf("[PRESENCE-SCHEDULER] %s available failed for %s: %v", label, device.Device, err)
				// If the device appears connected but SendPresence failed, it is a zombie:
				// socket is alive but the session is broken. Force a full reconnect.
				isConnected, _, _ := service.Status(ctx, device.Device)
				if isConnected && !isReconnectOnCooldown(device.Device) {
					logrus.Warnf("[ZOMBIE-DETECT] device %s connected but presence failed — forcing reconnect", device.Device)
					markReconnectAttempt(device.Device)
					if reconnErr := service.Reconnect(ctx, device.Device); reconnErr != nil {
						logrus.Warnf("[ZOMBIE-DETECT] reconnect failed for device %s: %v", device.Device, reconnErr)
					} else {
						logrus.Infof("[ZOMBIE-DETECT] device %s recovered via reconnect", device.Device)
					}
				}
				return
			}
			logrus.Infof("[PRESENCE-SCHEDULER] %s: %s online for %ds", label, device.Device, seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
			_ = service.SendDevicePresence(ctx, device.Device, "unavailable")
		}(fireAt, onlineSec, win.name)
	}
}

func MultipartFormFileHeaderToBytes(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	return fileBytes, nil
}
