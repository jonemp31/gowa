package helpers

import (
	"context"
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

func MultipartFormFileHeaderToBytes(fileHeader *multipart.FileHeader) []byte {
	file, _ := fileHeader.Open()
	defer file.Close()

	fileBytes := make([]byte, fileHeader.Size)
	_, _ = file.Read(fileBytes)

	return fileBytes
}
