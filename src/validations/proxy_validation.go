package validations

import (
	"context"
	"fmt"

	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateSetProxy(ctx context.Context, deviceID string, proxyType string, host string, port int) error {
	err := validation.ValidateWithContext(ctx, &deviceID, validation.Required.Error("device_id is required"))
	if err != nil {
		return pkgError.ValidationError(fmt.Sprintf("device_id: %s", err.Error()))
	}

	err = validation.ValidateWithContext(ctx, &proxyType,
		validation.Required.Error("type is required"),
		validation.In("socks5", "http", "https").Error("type must be one of: socks5, http, https"),
	)
	if err != nil {
		return pkgError.ValidationError(fmt.Sprintf("type(%s): %s", proxyType, err.Error()))
	}

	err = validation.ValidateWithContext(ctx, &host, validation.Required.Error("host is required"))
	if err != nil {
		return pkgError.ValidationError(fmt.Sprintf("host: %s", err.Error()))
	}

	err = validation.ValidateWithContext(ctx, &port,
		validation.Required.Error("port is required"),
		validation.Min(1).Error("port must be between 1 and 65535"),
		validation.Max(65535).Error("port must be between 1 and 65535"),
	)
	if err != nil {
		return pkgError.ValidationError(fmt.Sprintf("port(%d): %s", port, err.Error()))
	}

	return nil
}

func ValidateRemoveProxy(ctx context.Context, deviceID string) error {
	err := validation.ValidateWithContext(ctx, &deviceID, validation.Required.Error("device_id is required"))
	if err != nil {
		return pkgError.ValidationError(fmt.Sprintf("device_id: %s", err.Error()))
	}
	return nil
}

func ValidateTestProxy(ctx context.Context, deviceID string) error {
	return ValidateRemoveProxy(ctx, deviceID)
}
