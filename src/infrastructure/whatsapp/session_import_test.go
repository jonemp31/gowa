package whatsapp

import (
	"bytes"
	"context"
	"encoding/base64"
	"path/filepath"
	"testing"

	domainApp "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/app"
	sqlitedrv "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func b64Bytes(n int, fill byte) string {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = fill
	}
	return base64.StdEncoding.EncodeToString(buf)
}

func TestDecodeKey32(t *testing.T) {
	t.Run("valid 32 bytes", func(t *testing.T) {
		k, err := decodeKey32("f", b64Bytes(32, 0xAB))
		require.NoError(t, err)
		assert.Equal(t, bytes.Repeat([]byte{0xAB}, 32), k[:])
	})

	t.Run("strips 0x05 DjbType prefix from 33 bytes", func(t *testing.T) {
		raw := append([]byte{0x05}, bytes.Repeat([]byte{0x07}, 32)...)
		k, err := decodeKey32("f", base64.StdEncoding.EncodeToString(raw))
		require.NoError(t, err)
		assert.Equal(t, bytes.Repeat([]byte{0x07}, 32), k[:])
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		_, err := decodeKey32("f", b64Bytes(31, 0x01))
		assert.Error(t, err)
	})

	t.Run("rejects invalid base64", func(t *testing.T) {
		_, err := decodeKey32("f", "not base64 @@@")
		assert.Error(t, err)
	})

	t.Run("rejects empty", func(t *testing.T) {
		_, err := decodeKey32("f", "")
		assert.Error(t, err)
	})
}

func TestDecodeSig64(t *testing.T) {
	k, err := decodeSig64("sig", b64Bytes(64, 0x09))
	require.NoError(t, err)
	assert.Len(t, k[:], 64)

	_, err = decodeSig64("sig", b64Bytes(63, 0x09))
	assert.Error(t, err)
}

func TestParseBaileysJID(t *testing.T) {
	t.Run("user:device@server", func(t *testing.T) {
		jid, err := parseBaileysJID("5516988223583:9@s.whatsapp.net", false)
		require.NoError(t, err)
		assert.Equal(t, "5516988223583", jid.User)
		assert.Equal(t, uint16(9), jid.Device)
		assert.Equal(t, types.DefaultUserServer, jid.Server)
	})

	t.Run("lid server", func(t *testing.T) {
		jid, err := parseBaileysJID("235390699044948:9@lid", true)
		require.NoError(t, err)
		assert.Equal(t, "235390699044948", jid.User)
		assert.Equal(t, uint16(9), jid.Device)
		assert.Equal(t, types.HiddenUserServer, jid.Server)
	})

	t.Run("no device suffix defaults to 0", func(t *testing.T) {
		jid, err := parseBaileysJID("5516988223583@s.whatsapp.net", false)
		require.NoError(t, err)
		assert.Equal(t, uint16(0), jid.Device)
	})

	t.Run("empty user is rejected", func(t *testing.T) {
		_, err := parseBaileysJID("@s.whatsapp.net", false)
		assert.Error(t, err)
	})
}

func validTestCreds() domainApp.BaileysCreds {
	return domainApp.BaileysCreds{
		NoiseKey:          domainApp.BaileysKeyPair{Private: b64Bytes(32, 1), Public: b64Bytes(32, 2)},
		SignedIdentityKey: domainApp.BaileysKeyPair{Private: b64Bytes(32, 3), Public: b64Bytes(32, 4)},
		SignedPreKey: domainApp.BaileysSignedPreKey{
			KeyID:     1,
			KeyPair:   domainApp.BaileysKeyPair{Private: b64Bytes(32, 5), Public: b64Bytes(32, 6)},
			Signature: b64Bytes(64, 7),
		},
		RegistrationID: 12378,
		AdvSecretKey:   nil,
		Account: domainApp.BaileysAccount{
			Details:             base64.StdEncoding.EncodeToString([]byte{1, 2, 3}),
			AccountSignatureKey: b64Bytes(32, 8),
			AccountSignature:    b64Bytes(64, 9),
			DeviceSignature:     b64Bytes(64, 10),
		},
		Me:       domainApp.BaileysMe{ID: "5516988223583:9@s.whatsapp.net", LID: "235390699044948:9@lid"},
		Platform: "web",
	}
}

func newTestContainer(t *testing.T) *sqlstore.Container {
	t.Helper()
	dbPath := filepath.ToSlash(filepath.Join(t.TempDir(), "wa_test.db"))
	addr := sqlitedrv.FormatChatStorageURI("file:"+dbPath, true, true)
	c, err := sqlstore.New(context.Background(), sqlitedrv.DriverName, addr, waLog.Noop)
	require.NoError(t, err)
	// Close the DB before t.TempDir cleanup so Windows can unlink the file (LIFO: runs first).
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestBuildDeviceFromBaileysCreds(t *testing.T) {
	container := newTestContainer(t)

	t.Run("happy path maps all fields", func(t *testing.T) {
		dev, err := BuildDeviceFromBaileysCreds(container, validTestCreds())
		require.NoError(t, err)
		require.NotNil(t, dev.ID)
		assert.Equal(t, "5516988223583", dev.ID.User)
		assert.Equal(t, uint16(9), dev.ID.Device)
		assert.Equal(t, uint32(12378), dev.RegistrationID)
		assert.NotNil(t, dev.NoiseKey)
		assert.NotNil(t, dev.IdentityKey)
		assert.NotNil(t, dev.SignedPreKey)
		assert.NotNil(t, dev.Account)
		// advSecretKey was null -> NewDevice's random 32-byte value is kept (adv_key is NOT NULL).
		assert.Len(t, dev.AdvSecretKey, 32)
		// Initialized MUST remain false so PutDevice wires the sub-stores.
		assert.False(t, dev.Initialized)

		// It must persist and connect-wire without error.
		require.NoError(t, container.PutDevice(context.Background(), dev))
		assert.NotNil(t, dev.PreKeys, "PutDevice should have wired the PreKeys store")
	})

	t.Run("missing me.id is rejected", func(t *testing.T) {
		creds := validTestCreds()
		creds.Me.ID = ""
		_, err := BuildDeviceFromBaileysCreds(container, creds)
		assert.Error(t, err)
	})

	t.Run("malformed key length is rejected", func(t *testing.T) {
		creds := validTestCreds()
		creds.NoiseKey.Public = b64Bytes(16, 1)
		_, err := BuildDeviceFromBaileysCreds(container, creds)
		assert.Error(t, err)
	})

	t.Run("explicit advSecretKey is decoded", func(t *testing.T) {
		creds := validTestCreds()
		adv := b64Bytes(32, 0x5A)
		creds.AdvSecretKey = &adv
		dev, err := BuildDeviceFromBaileysCreds(container, creds)
		require.NoError(t, err)
		assert.Equal(t, bytes.Repeat([]byte{0x5A}, 32), dev.AdvSecretKey)
	})
}
