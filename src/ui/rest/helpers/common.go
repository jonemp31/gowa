package helpers

import (
	"context"
	"math/rand"
	"mime/multipart"
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
						logrus.Warnf("[AUTO-RECONNECT] failed for device %s: %v", device.Device, err)
					} else {
						logrus.Infof("[AUTO-RECONNECT] recovered device %s", device.Device)
					}
					time.Sleep(1000 * time.Millisecond)
				}
			}
		}
	}()
}

// SetDailyPresenceScheduler runs a background goroutine that sends a randomised
// presence signal for every registered device once per day:
//   - Between 06:00–10:00 local time: PresenceAvailable (brief online signal)
//   - Between 20:00–00:00 local time: PresenceUnavailable (offline signal)
//
// Each device gets its own random time within the window so that 115 devices
// do not all broadcast at exactly the same second. The scheduler recalculates
// at each local midnight, so it stays accurate across DST changes.
func SetDailyPresenceScheduler(service domainApp.IAppUsecase) {
	go func() {
		for {
			now := time.Now()
			scheduleDayPresences(service, now)
			// Sleep until 00:01 next day to recalculate for the new date
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

	// Morning window 06:00–10:00 → send PresenceAvailable briefly
	morningAt := today.Add(6*time.Hour + time.Duration(rand.Int63n(int64(4*time.Hour))))
	// Evening window 20:00–00:00 → send PresenceUnavailable
	eveningAt := today.Add(20*time.Hour + time.Duration(rand.Int63n(int64(4*time.Hour))))

	if morningAt.After(now) {
		time.Sleep(time.Until(morningAt))
		ctx := context.Background()
		if err := service.SendDevicePresence(ctx, device.Device, "available"); err != nil {
			logrus.Warnf("[PRESENCE-SCHEDULER] morning available failed for %s: %v", device.Device, err)
		} else {
			logrus.Infof("[PRESENCE-SCHEDULER] sent morning available for %s", device.Device)
		}
		// Brief online window then go back offline
		time.Sleep(time.Duration(10+rand.Intn(20)) * time.Second)
		_ = service.SendDevicePresence(ctx, device.Device, "unavailable")
	}

	if eveningAt.After(now) {
		time.Sleep(time.Until(eveningAt))
		ctx := context.Background()
		if err := service.SendDevicePresence(ctx, device.Device, "available"); err != nil {
			logrus.Warnf("[PRESENCE-SCHEDULER] evening available failed for %s: %v", device.Device, err)
		} else {
			logrus.Infof("[PRESENCE-SCHEDULER] sent evening available for %s", device.Device)
		}
		// Brief online window then go offline for the night
		time.Sleep(time.Duration(45+rand.Intn(46)) * time.Second)
		_ = service.SendDevicePresence(ctx, device.Device, "unavailable")
	}
}

func MultipartFormFileHeaderToBytes(fileHeader *multipart.FileHeader) []byte {
	file, _ := fileHeader.Open()
	defer file.Close()

	fileBytes := make([]byte, fileHeader.Size)
	_, _ = file.Read(fileBytes)

	return fileBytes
}
