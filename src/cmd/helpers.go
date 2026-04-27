package cmd

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/helpers"
	"github.com/sirupsen/logrus"
)

func startAutoReconnectCheckerIfClientAvailable() {
	if appUsecase == nil {
		logrus.Warn("app usecase is nil; auto-reconnect checker not started")
		return
	}
	helpers.SetAutoReconnectChecking(appUsecase)
}
