package validations

import (
	"context"

	domainUser "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/user"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateUserInfo(ctx context.Context, request domainUser.InfoRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Phone, validation.Required),
	)

	if err != nil {
		return pkgError.ValidationError(err.Error())
	}

	return nil
}
func ValidateUserAvatar(ctx context.Context, request domainUser.AvatarRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Phone, validation.Required),
		validation.Field(&request.IsCommunity, validation.When(request.IsCommunity, validation.Required, validation.In(true, false))),
		validation.Field(&request.IsPreview, validation.When(request.IsPreview, validation.Required, validation.In(true, false))),
	)

	if err != nil {
		return pkgError.ValidationError(err.Error())
	}

	return nil
}

func ValidateBusinessProfile(ctx context.Context, request domainUser.BusinessProfileRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Phone, validation.Required),
	)

	if err != nil {
		return pkgError.ValidationError(err.Error())
	}

	return nil
}

func ValidateSaveContact(ctx context.Context, request domainUser.SaveContactRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Phone, validation.Required),
		validation.Field(&request.FullName, validation.Required),
	)

	if err != nil {
		return pkgError.ValidationError(err.Error())
	}

	return nil
}

func ValidateSetStatusMessage(ctx context.Context, request domainUser.SetStatusMessageRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Message, validation.Required, validation.Length(1, 139)),
	)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}
	return nil
}

var validPrivacyNames = []interface{}{
	"groupadd", "last", "status", "profile", "readreceipts", "online", "calladd",
}

var validPrivacyValues = []interface{}{
	"all", "contacts", "contact_blacklist", "none", "match_last_seen", "known",
}

func ValidateSetPrivacySetting(ctx context.Context, request domainUser.SetPrivacySettingRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Name, validation.Required, validation.In(validPrivacyNames...)),
		validation.Field(&request.Value, validation.Required, validation.In(validPrivacyValues...)),
	)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}
	return nil
}

func ValidateSubscribePresence(ctx context.Context, request domainUser.SubscribePresenceRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Phone, validation.Required),
	)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}
	return nil
}
