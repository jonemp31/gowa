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

// fingerprintPool contains 500 realistic WhatsApp Web companion device fingerprints.
//
// Distribution (real-world WhatsApp Web usage approximation):
//
//	50% × Chrome  — Windows 10/11 · macOS Ventura/Sonoma/Sequoia · Linux x86_64
//	16% × Safari  — macOS Ventura/Sonoma/Sequoia
//	13% × Edge    — Windows 10/11 · macOS
//	13% × Desktop — Windows 10/11 · macOS
//	 6% × Firefox — Windows 10/11 · macOS · Linux x86_64
//	 2% × Opera   — Windows 10/11 · macOS
//
// Windows builds (6):
//
//	10.0.19044 Win10 21H2  ·  10.0.19045 Win10 22H2  ·  10.0.22000 Win11 21H2
//	10.0.22621 Win11 22H2  ·  10.0.22631 Win11 23H2  ·  10.0.26100 Win11 24H2
//
// macOS versions (30):
//
//	Ventura  13.x — 13.4.0 through 13.7.5  (9 releases)
//	Sonoma   14.x — 14.0.0 through 14.7.0  (11 releases)
//	Sequoia  15.x — 15.0.0 through 15.4.1  (10 releases)
//
// Distinct (Os, PlatformType) combinations: 150
// Total entries: 500
var fingerprintPool = []DeviceFingerprint{

	// ── Chrome · Windows (74 entries) ────────────────────────────────────────
	// Win10 21H2 · 10.0.19044 (6)
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Win10 22H2 · 10.0.19045 — last Win10 feature update (15)
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
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Win11 21H2 · 10.0.22000 — first Win11 (5)
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Win11 22H2 · 10.0.22621 (10)
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
	// Win11 23H2 · 10.0.22631 — dominant build in 2026 (21)
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
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_CHROME},
	// Win11 24H2 · 10.0.26100 — rapidly growing in 2026 (17)
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
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_CHROME},

	// ── Chrome · macOS Ventura 13.x (28 entries) ─────────────────────────────
	{Os: "Mac OS X 13.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.5.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.5.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.5", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.5", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_CHROME},

	// ── Chrome · macOS Sonoma 14.x (71 entries) ──────────────────────────────
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_CHROME},

	// ── Chrome · macOS Sequoia 15.x (74 entries) ─────────────────────────────
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_CHROME},

	// ── Chrome · Linux x86_64 (4 entries) ────────────────────────────────────
	{Os: "Linux x86_64", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Linux x86_64", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Linux x86_64", PlatformType: waCompanionReg.DeviceProps_CHROME},
	{Os: "Linux x86_64", PlatformType: waCompanionReg.DeviceProps_CHROME},

	// ── Safari · macOS Ventura 13.x (14 entries) ─────────────────────────────
	{Os: "Mac OS X 13.4.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.5.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.5.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.5.2", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.6.2", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.6.5", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_SAFARI},

	// ── Safari · macOS Sonoma 14.x (31 entries) ──────────────────────────────
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.2.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.2.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},

	// ── Safari · macOS Sequoia 15.x (34 entries) ─────────────────────────────
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.1.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.2.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.3.2", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},
	{Os: "Mac OS X 15.4.1", PlatformType: waCompanionReg.DeviceProps_SAFARI},

	// ── Edge · Windows (23 entries) ──────────────────────────────────────────
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_EDGE},

	// ── Edge · macOS (42 entries) ─────────────────────────────────────────────
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_EDGE},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_EDGE},

	// ── WhatsApp Desktop · Windows (23 entries) ───────────────────────────────
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_DESKTOP},

	// ── WhatsApp Desktop · macOS (42 entries) ────────────────────────────────
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.2.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.3.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.6.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.0.1", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.3.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_DESKTOP},

	// ── Firefox · Windows (11 entries) ───────────────────────────────────────
	{Os: "Windows 10.0.19044", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_FIREFOX},

	// ── Firefox · macOS (18 entries) ─────────────────────────────────────────
	{Os: "Mac OS X 13.5.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 13.5.2", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 13.6.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 13.6.2", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 13.6.7", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 13.7.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 13.7.5", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 14.0.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 14.1.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 14.3.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 14.4.1", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 14.6.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 15.1.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_FIREFOX},

	// ── Firefox · Linux x86_64 (1 entry) ─────────────────────────────────────
	{Os: "Linux x86_64", PlatformType: waCompanionReg.DeviceProps_FIREFOX},

	// ── Opera · Windows (5 entries) ──────────────────────────────────────────
	{Os: "Windows 10.0.19045", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Windows 10.0.22000", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Windows 10.0.22621", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Windows 10.0.22631", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Windows 10.0.26100", PlatformType: waCompanionReg.DeviceProps_OPERA},

	// ── Opera · macOS (5 entries) ─────────────────────────────────────────────
	{Os: "Mac OS X 14.5.0", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Mac OS X 14.7.0", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Mac OS X 15.0.0", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Mac OS X 15.2.0", PlatformType: waCompanionReg.DeviceProps_OPERA},
	{Os: "Mac OS X 15.4.0", PlatformType: waCompanionReg.DeviceProps_OPERA},
}
