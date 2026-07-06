package validations

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateLoginWithCode(ctx context.Context, phoneNumber string) error {
	// Combine validations using a single ValidateWithContext call
	err := validation.ValidateWithContext(ctx, &phoneNumber,
		validation.Required,
		validation.Match(regexp.MustCompile(`^\+?[0-9]{1,15}$`)),
	)
	if err != nil {
		return pkgError.ValidationError(fmt.Sprintf("phone_number(%s): %s", phoneNumber, err.Error()))
	}
	return nil
}

// ValidateImportSession checks that the Baileys credentials carry every field required
// to reconstruct a whatsmeow device. Byte-level decoding/length checks happen later in
// BuildDeviceFromBaileysCreds; here we only guard against missing pieces.
func ValidateImportSession(_ context.Context, creds domainApp.BaileysCreds) error {
	missing := func(name string) error {
		return pkgError.ValidationError(fmt.Sprintf("%s is required in session credentials", name))
	}

	if strings.TrimSpace(creds.Me.ID) == "" {
		return missing("me.id")
	}
	if strings.TrimSpace(creds.NoiseKey.Private) == "" || strings.TrimSpace(creds.NoiseKey.Public) == "" {
		return missing("noiseKey")
	}
	if strings.TrimSpace(creds.SignedIdentityKey.Private) == "" || strings.TrimSpace(creds.SignedIdentityKey.Public) == "" {
		return missing("signedIdentityKey")
	}
	if strings.TrimSpace(creds.SignedPreKey.KeyPair.Private) == "" ||
		strings.TrimSpace(creds.SignedPreKey.KeyPair.Public) == "" ||
		strings.TrimSpace(creds.SignedPreKey.Signature) == "" {
		return missing("signedPreKey")
	}
	if strings.TrimSpace(creds.Account.Details) == "" ||
		strings.TrimSpace(creds.Account.AccountSignature) == "" ||
		strings.TrimSpace(creds.Account.AccountSignatureKey) == "" ||
		strings.TrimSpace(creds.Account.DeviceSignature) == "" {
		return missing("account")
	}
	return nil
}
