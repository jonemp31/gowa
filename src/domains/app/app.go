package app

import (
	"context"
	"time"
)

type IAppUsecase interface {
	Login(ctx context.Context, deviceID string) (response LoginResponse, err error)
	LoginWithCode(ctx context.Context, deviceID string, phoneNumber string) (loginCode string, err error)
	ImportSession(ctx context.Context, deviceID string, creds BaileysCreds) (response ImportSessionResponse, err error)
	ImportPrivacyTokens(ctx context.Context, deviceID string, req ImportPrivacyTokensRequest) (response ImportPrivacyTokensResponse, err error)
	Logout(ctx context.Context, deviceID string) (err error)
	Reconnect(ctx context.Context, deviceID string) (err error)
	Status(ctx context.Context, deviceID string) (isConnected bool, isLoggedIn bool, err error)
	FirstDevice(ctx context.Context) (response DevicesResponse, err error)
	FetchDevices(ctx context.Context) (response []DevicesResponse, err error)
	SendDevicePresence(ctx context.Context, deviceID string, presence string) error
}

// BaileysCreds mirrors the JSON produced by the "Passkey Linker" browser extension
// (Baileys authentication credentials format), used to migrate an already-authenticated
// WhatsApp Web session into this API without a QR/passkey pairing flow.
type BaileysCreds struct {
	NoiseKey          BaileysKeyPair      `json:"noiseKey"`
	SignedIdentityKey BaileysKeyPair      `json:"signedIdentityKey"`
	SignedPreKey      BaileysSignedPreKey `json:"signedPreKey"`
	RegistrationID    uint32              `json:"registrationId"`
	AdvSecretKey      *string             `json:"advSecretKey"` // usually null (deleted post-pairing)
	Account           BaileysAccount      `json:"account"`
	Me                BaileysMe           `json:"me"`
	Platform          string              `json:"platform"`
}

type BaileysKeyPair struct {
	Private string `json:"private"`
	Public  string `json:"public"`
}

type BaileysSignedPreKey struct {
	KeyID     uint32         `json:"keyId"`
	KeyPair   BaileysKeyPair `json:"keyPair"`
	Signature string         `json:"signature"`
}

// BaileysAccount is the ADVSignedDeviceIdentity (device identity signed by the primary phone).
type BaileysAccount struct {
	Details             string `json:"details"`
	AccountSignatureKey string `json:"accountSignatureKey"`
	AccountSignature    string `json:"accountSignature"`
	DeviceSignature     string `json:"deviceSignature"`
}

type BaileysMe struct {
	ID   string  `json:"id"`  // "<user>:<device>@<server>"
	LID  string  `json:"lid"` // "<user>:<device>@lid"
	Name *string `json:"name"`
}

type ImportSessionResponse struct {
	DeviceID   string `json:"device_id"`
	JID        string `json:"jid"`
	IsLoggedIn bool   `json:"is_logged_in"`
}

// PrivacyTokenEntry is one trusted-contact (tc) token extracted from the browser's
// in-memory chat model, used to pre-warm a migrated session so it can message existing
// contacts without WhatsApp error 463.
type PrivacyTokenEntry struct {
	LID                    string `json:"lid"`   // "<user>@lid"
	Phone                  string `json:"phone"` // phone number (digits) or "<user>@..."
	Token                  string `json:"token"` // base64-encoded tc token
	TcTokenTimestamp       int64  `json:"tc_token_timestamp"`
	TcTokenSenderTimestamp int64  `json:"tc_token_sender_timestamp"`
}

type ImportPrivacyTokensRequest struct {
	Tokens []PrivacyTokenEntry `json:"tokens"`
}

type ImportPrivacyTokensResponse struct {
	DeviceID string `json:"device_id"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
}

type DevicesResponse struct {
	Name      string    `json:"name"`
	Device    string    `json:"device"`
	JID       string    `json:"jid"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	ImagePath string        `json:"image_path"`
	Duration  time.Duration `json:"duration"`
	Code      string        `json:"code"`
}
