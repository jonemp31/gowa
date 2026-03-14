package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/device"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"github.com/gofiber/fiber/v2"
)

type Proxy struct {
	Service device.IDeviceUsecase
}

func InitRestProxy(app fiber.Router, service device.IDeviceUsecase) Proxy {
	rest := Proxy{Service: service}

	app.Put("/proxy", rest.SetProxy)
	app.Delete("/proxy", rest.RemoveProxy)
	app.Get("/proxy/test", rest.TestProxy)

	return rest
}

type setProxyRequest struct {
	DeviceID string `json:"device_id"`
	Type     string `json:"type"`     // socks5, http, https
	Host     string `json:"host"`     // IP or hostname
	Port     int    `json:"port"`     // 1-65535
	Username string `json:"username"` // optional
	Password string `json:"password"` // optional
}

func (handler *Proxy) SetProxy(c *fiber.Ctx) error {
	var req setProxyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}

	if err := validations.ValidateSetProxy(c.UserContext(), req.DeviceID, req.Type, req.Host, req.Port); err != nil {
		utils.PanicIfNeeded(err)
	}

	err := handler.Service.SetDeviceProxy(c.UserContext(), req.DeviceID, req.Type, req.Host, req.Port, req.Username, req.Password)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Proxy set successfully",
		Results: map[string]any{
			"device_id": req.DeviceID,
			"proxy":     whatsapp.MaskProxyURL(req.Type + "://" + req.Host),
		},
	})
}

type removeProxyRequest struct {
	DeviceID string `json:"device_id"`
}

func (handler *Proxy) RemoveProxy(c *fiber.Ctx) error {
	var req removeProxyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}

	if err := validations.ValidateRemoveProxy(c.UserContext(), req.DeviceID); err != nil {
		utils.PanicIfNeeded(err)
	}

	err := handler.Service.RemoveDeviceProxy(c.UserContext(), req.DeviceID)
	utils.PanicIfNeeded(err)

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Proxy removed successfully",
		Results: map[string]any{
			"device_id": req.DeviceID,
		},
	})
}

func (handler *Proxy) TestProxy(c *fiber.Ctx) error {
	deviceID := c.Query("device_id")

	if err := validations.ValidateTestProxy(c.UserContext(), deviceID); err != nil {
		utils.PanicIfNeeded(err)
	}

	healthy, externalIP, err := handler.Service.TestDeviceProxy(c.UserContext(), deviceID)
	utils.PanicIfNeeded(err)

	// Get masked proxy URL for safe display
	maskedProxy := ""
	if inst, ok := whatsapp.GetDeviceManager().GetDevice(deviceID); ok {
		maskedProxy = whatsapp.MaskProxyURL(inst.ProxyURL())
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Proxy test completed",
		Results: map[string]any{
			"device_id":   deviceID,
			"proxy":       maskedProxy,
			"healthy":     healthy,
			"external_ip": externalIP,
		},
	})
}
