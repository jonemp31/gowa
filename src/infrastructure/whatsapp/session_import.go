package whatsapp

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"

	waAdv "go.mau.fi/whatsmeow/proto/waAdv"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/util/keys"
)

// BuildDeviceFromBaileysCreds converts the Baileys-format credentials exported by the
// browser extension into a whatsmeow *store.Device ready to be persisted and connected.
//
// It relies on container.NewDevice() to wire the sub-stores (Identities/Sessions/PreKeys/…)
// via the Container; the crypto material is then overwritten with the imported values.
//
// IMPORTANT: the returned device MUST NOT have Initialized=true set before PutDevice —
// PutDevice only wires the sub-stores (initializeDevice) when Initialized == false, so
// setting it true results in a nil PreKeys store and a panic on connect.
func BuildDeviceFromBaileysCreds(container *sqlstore.Container, creds domainApp.BaileysCreds) (*store.Device, error) {
	if container == nil {
		return nil, fmt.Errorf("nil store container")
	}

	dev := container.NewDevice()

	nkPub, err := decodeKey32("noiseKey.public", creds.NoiseKey.Public)
	if err != nil {
		return nil, err
	}
	nkPriv, err := decodeKey32("noiseKey.private", creds.NoiseKey.Private)
	if err != nil {
		return nil, err
	}
	dev.NoiseKey = &keys.KeyPair{Pub: nkPub, Priv: nkPriv}

	idPub, err := decodeKey32("signedIdentityKey.public", creds.SignedIdentityKey.Public)
	if err != nil {
		return nil, err
	}
	idPriv, err := decodeKey32("signedIdentityKey.private", creds.SignedIdentityKey.Private)
	if err != nil {
		return nil, err
	}
	dev.IdentityKey = &keys.KeyPair{Pub: idPub, Priv: idPriv}

	spkPub, err := decodeKey32("signedPreKey.keyPair.public", creds.SignedPreKey.KeyPair.Public)
	if err != nil {
		return nil, err
	}
	spkPriv, err := decodeKey32("signedPreKey.keyPair.private", creds.SignedPreKey.KeyPair.Private)
	if err != nil {
		return nil, err
	}
	spkSig, err := decodeSig64("signedPreKey.signature", creds.SignedPreKey.Signature)
	if err != nil {
		return nil, err
	}
	dev.SignedPreKey = &keys.PreKey{
		KeyPair:   keys.KeyPair{Pub: spkPub, Priv: spkPriv},
		KeyID:     creds.SignedPreKey.KeyID,
		Signature: spkSig,
	}

	dev.RegistrationID = creds.RegistrationID

	// advSecretKey is usually null (deleted by WhatsApp Web post-pairing). The DB column
	// adv_key is NOT NULL, and reconnection of an already-paired device does not use the
	// real secret — so when absent we keep the random value NewDevice() already generated.
	if creds.AdvSecretKey != nil && strings.TrimSpace(*creds.AdvSecretKey) != "" {
		adv, err := decodeBytes("advSecretKey", *creds.AdvSecretKey)
		if err != nil {
			return nil, err
		}
		dev.AdvSecretKey = adv
	}

	details, err := decodeBytes("account.details", creds.Account.Details)
	if err != nil {
		return nil, err
	}
	accSigKey, err := decodeBytes("account.accountSignatureKey", creds.Account.AccountSignatureKey)
	if err != nil {
		return nil, err
	}
	accSig, err := decodeBytes("account.accountSignature", creds.Account.AccountSignature)
	if err != nil {
		return nil, err
	}
	devSig, err := decodeBytes("account.deviceSignature", creds.Account.DeviceSignature)
	if err != nil {
		return nil, err
	}
	dev.Account = &waAdv.ADVSignedDeviceIdentity{
		Details:             details,
		AccountSignatureKey: accSigKey,
		AccountSignature:    accSig,
		DeviceSignature:     devSig,
	}

	if strings.TrimSpace(creds.Me.ID) == "" {
		return nil, fmt.Errorf("me.id is required")
	}
	jid, err := parseBaileysJID(creds.Me.ID, false)
	if err != nil {
		return nil, fmt.Errorf("me.id: %w", err)
	}
	dev.ID = &jid

	if strings.TrimSpace(creds.Me.LID) != "" {
		if lid, err := parseBaileysJID(creds.Me.LID, true); err == nil {
			dev.LID = lid
		}
	}

	if creds.Me.Name != nil {
		dev.PushName = *creds.Me.Name
	}
	dev.Platform = creds.Platform

	// NOTE: intentionally NOT setting dev.Initialized — see the doc comment above.
	return dev, nil
}

// decodeB64 decodes standard base64, falling back to raw (unpadded) base64.
func decodeB64(field, s string) ([]byte, error) {
	if strings.TrimSpace(s) == "" {
		return nil, fmt.Errorf("%s: empty", field)
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	b, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid base64: %w", field, err)
	}
	return b, nil
}

// decodeBytes returns the raw decoded bytes (opaque values: account fields, advSecret).
func decodeBytes(field, s string) ([]byte, error) {
	return decodeB64(field, s)
}

// decodeKey32 decodes a 32-byte Curve25519 key, stripping the libsignal DjbType (0x05)
// prefix if the value arrives as 33 bytes.
func decodeKey32(field, s string) (*[32]byte, error) {
	b, err := decodeB64(field, s)
	if err != nil {
		return nil, err
	}
	if len(b) == 33 && b[0] == 0x05 {
		b = b[1:]
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("%s: expected 32-byte key, got %d", field, len(b))
	}
	var a [32]byte
	copy(a[:], b)
	return &a, nil
}

// decodeSig64 decodes a 64-byte signature.
func decodeSig64(field, s string) (*[64]byte, error) {
	b, err := decodeB64(field, s)
	if err != nil {
		return nil, err
	}
	if len(b) != 64 {
		return nil, fmt.Errorf("%s: expected 64-byte signature, got %d", field, len(b))
	}
	var a [64]byte
	copy(a[:], b)
	return &a, nil
}

// parseBaileysJID parses the Baileys JID form "<user>:<device>@<server>" into a types.JID.
// whatsmeow's ParseJID expects a different layout, so we build it field-by-field.
func parseBaileysJID(s string, hidden bool) (types.JID, error) {
	server := types.DefaultUserServer
	if hidden {
		server = types.HiddenUserServer
	}
	head := s
	if at := strings.LastIndex(s, "@"); at >= 0 {
		head = s[:at]
		server = s[at+1:]
	}
	if strings.TrimSpace(head) == "" {
		return types.JID{}, fmt.Errorf("empty user in JID %q", s)
	}
	user := head
	var device uint16
	if colon := strings.Index(head, ":"); colon >= 0 {
		user = head[:colon]
		if d, err := strconv.ParseUint(head[colon+1:], 10, 16); err == nil {
			device = uint16(d)
		}
	}
	return types.JID{User: user, Device: device, Server: server}, nil
}
