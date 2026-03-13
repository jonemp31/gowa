package whatsapp

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
)

// DeviceFingerprint represents a WhatsApp companion device identity
// consisting of an OS string and a platform type.
type DeviceFingerprint struct {
	Os           string
	PlatformType waCompanionReg.DeviceProps_PlatformType
}

// String encodes the fingerprint as "platformType|os" for DB storage.
func (fp DeviceFingerprint) String() string {
	return fmt.Sprintf("%d|%s", int32(fp.PlatformType), fp.Os)
}

// ParseFingerprint decodes a fingerprint from its string representation.
func ParseFingerprint(s string) (DeviceFingerprint, bool) {
	parts := strings.SplitN(s, "|", 2)
	if len(parts) != 2 {
		return DeviceFingerprint{}, false
	}
	pt, err := strconv.Atoi(parts[0])
	if err != nil {
		return DeviceFingerprint{}, false
	}
	return DeviceFingerprint{
		Os:           parts[1],
		PlatformType: waCompanionReg.DeviceProps_PlatformType(pt),
	}, true
}

// devicePropsMu protects store.DeviceProps during concurrent client creation.
// Must be held from applyFingerprint through whatsmeow.NewClient to prevent
// race conditions when multiple devices are initialised in parallel.
var devicePropsMu sync.Mutex

// applyFingerprintLocked sets the global store.DeviceProps.
// The caller MUST hold devicePropsMu.
func applyFingerprintLocked(fp DeviceFingerprint) {
	pt := fp.PlatformType
	os := fp.Os
	store.DeviceProps.PlatformType = &pt
	store.DeviceProps.Os = &os
}

func init() {
	if len(fingerprintPool) == 0 {
		panic("fingerprintPool must not be empty")
	}
}

// RandomFingerprint returns a randomly chosen fingerprint from the pool.
func RandomFingerprint() DeviceFingerprint {
	return fingerprintPool[rand.Intn(len(fingerprintPool))]
}

// fingerprintPool contains 100 realistic WhatsApp Web companion device fingerprints.
//
// Distribution (matches real-world WhatsApp Web usage):
//
//	60 × Chrome on Windows   (PlatformType = CHROME  = 1)
//	15 × Chrome on macOS     (PlatformType = CHROME  = 1)
//	10 × Edge on Windows     (PlatformType = EDGE    = 6)
//	10 × WhatsApp Desktop    (PlatformType = DESKTOP = 7)
//	 2 × Firefox on Windows  (PlatformType = FIREFOX = 2)
//	 3 × Safari on macOS     (PlatformType = SAFARI  = 5)
var fingerprintPool = []DeviceFingerprint{
	// ── Chrome on Windows (60) ──────────────────────────────────────────
	// Windows 10 22H2 (build 19045) — most popular Win10 build
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Windows 10 21H2 (build 19044)
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Windows 11 22H2 (build 22621)
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Windows 11 23H2 (build 22631) — most popular Win11 build
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Windows 11 24H2 (build 26100)
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Windows 10 1909 (build 18363) — older but still in use
	{Os: "Windows 10.0.18363", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.18363", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Windows 10 21H1 (build 19043)
	{Os: "Windows 10.0.19043", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19043", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Windows 11 21H2 (build 22000) — first Win11
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_CHROME},

	// ── Chrome on macOS (15) ────────────────────────────────────────────
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},

	// ── Edge on Windows (10) ────────────────────────────────────────────
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_EDGE},

	// ── WhatsApp Desktop on Windows (10) ────────────────────────────────
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_DESKTOP},

	// ── Firefox on Windows (2) ──────────────────────────────────────────
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_FIREFOX},

	// ── Safari on macOS (3) ─────────────────────────────────────────────
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
}
