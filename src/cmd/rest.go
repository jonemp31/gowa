package cmd

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/helpers"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/middleware"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/websocket"
	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Send whatsapp API over http",
	Long:  `This application is from clone https://github.com/aldinokemal/go-whatsapp-web-multidevice`,
	Run:   restServer,
}

func init() {
	rootCmd.AddCommand(restCmd)
}
func restServer(_ *cobra.Command, _ []string) {
	engine := html.NewFileSystem(http.FS(EmbedIndex), ".html")
	engine.AddFunc("isEnableBasicAuth", func(token any) bool {
		return token != nil
	})
	fiberConfig := fiber.Config{
		Views:                   engine,
		EnableTrustedProxyCheck: true,
		BodyLimit:               int(config.WhatsappSettingMaxVideoSize),
		Network:                 "tcp",
	}

	// Configure proxy settings if trusted proxies are specified
	if len(config.AppTrustedProxies) > 0 {
		fiberConfig.TrustedProxies = config.AppTrustedProxies
		fiberConfig.ProxyHeader = fiber.HeaderXForwardedHost
	}

	app := fiber.New(fiberConfig)

	app.Static(config.AppBasePath+"/statics", "./statics")
	app.Use(config.AppBasePath+"/components", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/components",
		Browse:     true,
	}))
	app.Use(config.AppBasePath+"/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/assets",
		Browse:     true,
	}))

	app.Use(middleware.Recovery())
	requestTimeout := time.Duration(config.AppRequestTimeoutSeconds) * time.Second
	if requestTimeout <= 0 {
		requestTimeout = middleware.DefaultRequestTimeout
	}
	app.Use(middleware.RequestTimeout(requestTimeout))
	app.Use(middleware.BasicAuth())
	app.Use(logger.New(logger.Config{
		Format:     "${time} [HTTP]  ${method} ${path} | ${status} | ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Device manager - needed for chatwoot webhook and health check
	dm := whatsapp.GetDeviceManager()

	// Health check endpoint — public, no auth.
	// Registered at root (ignoring AppBasePath) so infrastructure probes (Docker
	// healthcheck, load-balancer, Swarm) always find it at a fixed path.
	//
	// Status rules:
	//   healthy   — >95 % of devices are logged_in or connected
	//   degraded  — 80–95 % connected
	//   unhealthy — <80 % connected  OR device manager not initialised
	app.Get("/health", func(c *fiber.Ctx) error {
		type deviceCounts struct {
			Total        int `json:"total"`
			LoggedIn     int `json:"logged_in"`
			Connected    int `json:"connected"`
			Disconnected int `json:"disconnected"`
		}
		type healthResponse struct {
			Status                        string       `json:"status"`
			Devices                       deviceCounts `json:"devices"`
			Goroutines                    int          `json:"goroutines"`
			WebhookSemaphoreUsagePct      int          `json:"webhook_semaphore_usage_pct"`
			WebhookDispatchSemaphoreUsage int          `json:"webhook_dispatch_semaphore_usage_pct"`
			UptimeSeconds                 int64        `json:"uptime_seconds"`
			Version                       string       `json:"version"`
		}

		sem := whatsapp.GetWebhookSemaphoreStats()

		resp := healthResponse{
			Goroutines:                    runtime.NumGoroutine(),
			WebhookSemaphoreUsagePct:      sem.DeliveryUsagePct,
			WebhookDispatchSemaphoreUsage: sem.DispatchUsagePct,
			UptimeSeconds:                 whatsapp.GetUptimeSeconds(),
			Version:                       config.AppVersion,
		}

		if dm == nil || !dm.IsHealthy() {
			resp.Status = "unhealthy"
			return c.Status(http.StatusServiceUnavailable).JSON(resp)
		}

		counts := dm.DeviceCountsByState()
		resp.Devices = deviceCounts{
			Total:        counts.Total,
			LoggedIn:     counts.LoggedIn,
			Connected:    counts.Connected,
			Disconnected: counts.Disconnected,
		}

		functional := counts.LoggedIn + counts.Connected
		var pct int
		if counts.Total > 0 {
			pct = functional * 100 / counts.Total
		} else {
			// No devices registered yet — manager is healthy, nothing to connect.
			pct = 100
		}

		switch {
		case pct >= 95:
			resp.Status = "healthy"
			return c.JSON(resp)
		case pct >= 80:
			resp.Status = "degraded"
			return c.Status(http.StatusOK).JSON(resp)
		default:
			resp.Status = "unhealthy"
			return c.Status(http.StatusServiceUnavailable).JSON(resp)
		}
	})

	// Chatwoot webhook - registered BEFORE basic auth middleware
	// This allows Chatwoot to send webhooks without authentication
	if config.ChatwootEnabled {
		chatwootHandler := rest.NewChatwootHandler(appUsecase, sendUsecase, dm, chatStorageRepo)
		webhookPath := "/chatwoot/webhook"
		if config.AppBasePath != "" {
			webhookPath = config.AppBasePath + webhookPath
		}
		app.Post(webhookPath, chatwootHandler.HandleWebhook)
	}

	if len(config.AppBasicAuthCredential) > 0 {
		account := make(map[string]string)
		for _, basicAuth := range config.AppBasicAuthCredential {
			ba := strings.Split(basicAuth, ":")
			if len(ba) != 2 {
				logrus.Fatalln("Basic auth is not valid, please this following format <user>:<secret>")
			}
			account[ba[0]] = ba[1]
		}

		app.Use(basicauth.New(basicauth.Config{
			Users: account,
		}))
	}

	// Create base path group or use app directly
	var apiGroup fiber.Router = app
	if config.AppBasePath != "" {
		apiGroup = app.Group(config.AppBasePath)
	}

	registerDeviceScopedRoutes := func(r fiber.Router) {
		rest.InitRestApp(r, appUsecase)
		rest.InitRestChat(r, chatUsecase)
		rest.InitRestSend(r, sendUsecase)
		rest.InitRestUser(r, userUsecase)
		rest.InitRestMessage(r, messageUsecase)
		rest.InitRestGroup(r, groupUsecase)
		rest.InitRestNewsletter(r, newsletterUsecase)
		websocket.RegisterRoutes(r, appUsecase)
	}

	// Device management routes (no device_id required)
	rest.InitRestDevice(apiGroup, deviceUsecase)
	rest.InitRestProxy(apiGroup, deviceUsecase)

	// Device-scoped operations (header-based)
	headerDeviceGroup := apiGroup.Group("", middleware.DeviceMiddleware(dm))
	registerDeviceScopedRoutes(headerDeviceGroup)

	// Chatwoot sync routes - require authentication (webhook is registered earlier without auth)
	if config.ChatwootEnabled {
		chatwootHandler := rest.NewChatwootHandler(appUsecase, sendUsecase, dm, chatStorageRepo)
		apiGroup.Post("/chatwoot/sync", chatwootHandler.SyncHistory)
		apiGroup.Get("/chatwoot/sync/status", chatwootHandler.SyncStatus)
	}

	apiGroup.Get("/", func(c *fiber.Ctx) error {
		return c.Render("views/index", fiber.Map{
			"AppHost":        fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname()),
			"AppVersion":     config.AppVersion,
			"AppBasePath":    config.AppBasePath,
			"BasicAuthToken": c.UserContext().Value(middleware.AuthorizationValue("BASIC_AUTH")),
			"MaxFileSize":    humanize.Bytes(uint64(config.WhatsappSettingMaxFileSize)),
			"MaxVideoSize":   humanize.Bytes(uint64(config.WhatsappSettingMaxVideoSize)),
		})
	})

	go websocket.RunHub()

	// Set auto reconnect to whatsapp server after booting
	go helpers.SetAutoConnectAfterBooting(appUsecase)

	// Set auto reconnect checking with a guaranteed client instance
	startAutoReconnectCheckerIfClientAvailable()

	// Human-like daily presence scheduler: 8 randomised windows per device per day
	go helpers.SetDailyPresenceScheduler(appUsecase)

	// Set daily presence pulse scheduler when enabled
	startPresencePulseSchedulerIfEnabled()

	if config.MediaRetentionDays > 0 {
		go func() {
			logrus.Infof("[MEDIA-CLEANUP] Media retention enabled: %d days, checking every 6 hours", config.MediaRetentionDays)
			whatsapp.CleanupOldMedia()
			ticker := time.NewTicker(6 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				whatsapp.CleanupOldMedia()
			}
		}()
	}

	if err := app.Listen(config.AppHost + ":" + config.AppPort); err != nil {
		logrus.Fatalln("Failed to start: ", err.Error())
	}
}
