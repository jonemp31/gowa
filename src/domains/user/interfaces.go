package user

import (
	"context"
)

// IUserInfo handles user information operations
type IUserInfo interface {
	Info(ctx context.Context, request InfoRequest) (response InfoResponse, err error)
	IsOnWhatsApp(ctx context.Context, request CheckRequest) (response CheckResponse, err error)
	BusinessProfile(ctx context.Context, request BusinessProfileRequest) (response BusinessProfileResponse, err error)
}

// IUserProfile handles user profile operations
type IUserProfile interface {
	Avatar(ctx context.Context, request AvatarRequest) (response AvatarResponse, err error)
	ChangeAvatar(ctx context.Context, request ChangeAvatarRequest) (err error)
	ChangePushName(ctx context.Context, request ChangePushNameRequest) (err error)
	SetStatusMessage(ctx context.Context, request SetStatusMessageRequest) (err error)
}

// IUserContact handles contact management operations
type IUserContact interface {
	SaveContact(ctx context.Context, request SaveContactRequest) (response SaveContactResponse, err error)
}

// IUserListing handles user listing operations
type IUserListing interface {
	MyListGroups(ctx context.Context) (response MyListGroupsResponse, err error)
	MyListNewsletter(ctx context.Context) (response MyListNewsletterResponse, err error)
	MyListContacts(ctx context.Context) (response MyListContactsResponse, err error)
}

// IUserPrivacy handles user privacy operations
type IUserPrivacy interface {
	MyPrivacySetting(ctx context.Context) (response MyPrivacySettingResponse, err error)
	SetPrivacySetting(ctx context.Context, request SetPrivacySettingRequest) (response MyPrivacySettingResponse, err error)
}

// IUserPresence handles presence subscription operations
type IUserPresence interface {
	SubscribePresence(ctx context.Context, request SubscribePresenceRequest) (response SubscribePresenceResponse, err error)
	SetForceActiveDeliveryReceipts(ctx context.Context, request SetForceActiveDeliveryReceiptsRequest) (err error)
}

// IUserUsecase combines all user interfaces for backward compatibility
type IUserUsecase interface {
	IUserInfo
	IUserProfile
	IUserListing
	IUserPrivacy
	IUserContact
	IUserPresence
}
