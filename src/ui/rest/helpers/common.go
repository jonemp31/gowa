package helpers

import (
	"context"
	"math/rand"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	"github.com/sirupsen/logrus"
)

func SetAutoConnectAfterBooting(service domainApp.IAppUsecase) {
	time.Sleep(2 * time.Second)
	devices, err := service.FetchDevices(context.Background())
	if err != nil || len(devices) == 0 {
		logrus.Warn("auto-connect skipped: no devices available")
		return
	}

	const maxWorkers = 5
	const launchDelay = 1200 * time.Millisecond

	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

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
				logrus.Warnf("[AUTO-CONNECT] failed for device %s: %v", d.Device, err)
			} else {
				logrus.Infof("[AUTO-CONNECT] connected device %s", d.Device)
			}
		}(device)
	}

	wg.Wait()
}

func SetAutoReconnectChecking(service domainApp.IAppUsecase) {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			devices, err := service.FetchDevices(context.Background())
			if err != nil || len(devices) == 0 {
				continue
			}
			for _, device := range devices {
				isConnected, _, _ := service.Status(context.Background(), device.Device)
				if !isConnected {
					if err := service.Reconnect(context.Background(), device.Device); err != nil {
						// Session deleted means the user revoked access from their phone.
						// The device can never reconnect — purge it so it stops generating
						// reconnect noise and triggers a device_removed webhook so downstream
						// systems (e.g. Supabase via n8n) can update their state.
						if strings.Contains(err.Error(), "session deleted") {
							logrus.Infof("[AUTO-RECONNECT] session deleted for device %s — removing", device.Device)
							if purgeErr := service.Logout(context.Background(), device.Device); purgeErr != nil {
								logrus.Warnf("[AUTO-RECONNECT] failed to remove session-deleted device %s: %v", device.Device, purgeErr)
							}
						} else {
							logrus.Warnf("[AUTO-RECONNECT] failed for device %s: %v", device.Device, err)
						}
					} else {
						logrus.Infof("[AUTO-RECONNECT] recovered device %s", device.Device)
					}
					time.Sleep(1000 * time.Millisecond)
				}
			}
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
				return
			}
			logrus.Infof("[PRESENCE-SCHEDULER] %s: %s online for %ds", label, device.Device, seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
			_ = service.SendDevicePresence(ctx, device.Device, "unavailable")
		}(fireAt, onlineSec, win.name)
	}
}

func MultipartFormFileHeaderToBytes(fileHeader *multipart.FileHeader) []byte {
	file, _ := fileHeader.Open()
	defer file.Close()

	fileBytes := make([]byte, fileHeader.Size)
	_, _ = file.Read(fileBytes)

	return fileBytes
}
