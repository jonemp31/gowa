package whatsapp

import (
	"math/rand"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

// proxyState tracks health status for a single proxy URL.
type proxyState struct {
	URL              string
	Healthy          bool
	ConsecutiveFails int
	LastChecked      time.Time
	LastLoggedDown   time.Time
}

// ProxyManager manages a pool of proxies with health checking and balanced assignment.
type ProxyManager struct {
	mu      sync.RWMutex
	proxies []*proxyState
	stopCh  chan struct{}
}

var (
	globalProxyManager *ProxyManager
	proxyManagerOnce   sync.Once
)

// InitProxyManager initializes the global proxy manager from config.
// Call this once at startup after config is loaded.
func InitProxyManager() {
	proxyManagerOnce.Do(func() {
		urls := config.WhatsappProxies
		if len(urls) == 0 {
			logrus.Info("[PROXY] No proxies configured — all devices will use direct connection (VPS IP)")
			return
		}

		pm := &ProxyManager{
			proxies: make([]*proxyState, 0, len(urls)),
			stopCh:  make(chan struct{}),
		}

		for _, u := range urls {
			if u == "" {
				continue
			}
			pm.proxies = append(pm.proxies, &proxyState{
				URL:     u,
				Healthy: true, // assume healthy until first check
			})
		}

		if len(pm.proxies) == 0 {
			logrus.Info("[PROXY] No valid proxies after parsing — using direct connection")
			return
		}

		logrus.Infof("[PROXY] Initialized proxy pool with %d proxies", len(pm.proxies))
		for i, p := range pm.proxies {
			logrus.Infof("[PROXY]   #%d: %s", i+1, maskProxyURL(p.URL))
		}

		globalProxyManager = pm
		go pm.healthLoop()
	})
}

// GetProxyManager returns the global proxy manager (may be nil if no proxies configured).
func GetProxyManager() *ProxyManager {
	return globalProxyManager
}

// SelectProxy picks a healthy proxy with the fewest assigned devices (balanced).
// deviceProxyCounts maps proxy URL → number of devices currently using it.
// Returns "" if no proxies are available (use direct connection).
func (pm *ProxyManager) SelectProxy(deviceProxyCounts map[string]int) string {
	if pm == nil {
		return ""
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var best *proxyState
	bestCount := int(^uint(0) >> 1) // max int

	// Collect healthy proxies with least usage
	var candidates []*proxyState
	for _, p := range pm.proxies {
		if !p.Healthy {
			continue
		}
		count := deviceProxyCounts[p.URL]
		if count < bestCount {
			bestCount = count
			candidates = []*proxyState{p}
		} else if count == bestCount {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		logrus.Warn("[PROXY] No healthy proxies available — using direct connection")
		return ""
	}

	// Random pick among equally-loaded candidates
	best = candidates[rand.Intn(len(candidates))]
	logrus.Infof("[PROXY] Selected proxy %s (load: %d devices)", maskProxyURL(best.URL), bestCount)
	return best.URL
}

// IsHealthy checks if a specific proxy URL is currently healthy.
func (pm *ProxyManager) IsHealthy(proxyURL string) bool {
	if pm == nil || proxyURL == "" {
		return true // no proxy = direct = always healthy
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, p := range pm.proxies {
		if p.URL == proxyURL {
			return p.Healthy
		}
	}
	return false
}

// FindHealthyReplacement returns a healthy proxy URL different from the given one,
// preferring the least-loaded. Returns "" if none available.
func (pm *ProxyManager) FindHealthyReplacement(currentProxy string, deviceProxyCounts map[string]int) string {
	if pm == nil {
		return ""
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var best *proxyState
	bestCount := int(^uint(0) >> 1)

	for _, p := range pm.proxies {
		if !p.Healthy || p.URL == currentProxy {
			continue
		}
		count := deviceProxyCounts[p.URL]
		if count < bestCount {
			bestCount = count
			best = p
		}
	}

	if best == nil {
		return ""
	}
	return best.URL
}

// ProxyURLs returns all configured proxy URLs.
func (pm *ProxyManager) ProxyURLs() []string {
	if pm == nil {
		return nil
	}
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	urls := make([]string, len(pm.proxies))
	for i, p := range pm.proxies {
		urls[i] = p.URL
	}
	return urls
}

// Stop terminates the health check loop.
func (pm *ProxyManager) Stop() {
	if pm == nil {
		return
	}
	close(pm.stopCh)
}

// healthLoop runs periodic health checks on all proxies.
func (pm *ProxyManager) healthLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run first check immediately
	pm.checkAll()

	for {
		select {
		case <-pm.stopCh:
			logrus.Info("[PROXY] Health monitor stopped")
			return
		case <-ticker.C:
			pm.checkAll()
		}
	}
}

func (pm *ProxyManager) checkAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, p := range pm.proxies {
		healthy := checkProxy(p.URL)
		p.LastChecked = time.Now()

		if healthy {
			if !p.Healthy {
				logrus.Infof("[PROXY] ✓ Proxy %s is back UP", maskProxyURL(p.URL))
			}
			p.Healthy = true
			p.ConsecutiveFails = 0
			p.LastLoggedDown = time.Time{}
		} else {
			p.ConsecutiveFails++
			if p.ConsecutiveFails >= 2 && p.Healthy {
				p.Healthy = false
				logrus.Errorf("[PROXY] ✗ Proxy %s is DOWN (failed %d consecutive checks)", maskProxyURL(p.URL), p.ConsecutiveFails)
				p.LastLoggedDown = time.Now()
			} else if !p.Healthy {
				// Re-log every 15 minutes while still down
				if time.Since(p.LastLoggedDown) >= 15*time.Minute {
					logrus.Errorf("[PROXY] ✗ Proxy %s still DOWN (failed %d checks)", maskProxyURL(p.URL), p.ConsecutiveFails)
					p.LastLoggedDown = time.Now()
				}
			}
		}
	}
}

// checkProxy tests TCP connectivity to a proxy.
func checkProxy(proxyURL string) bool {
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return false
	}

	host := parsed.Host
	if host == "" {
		return false
	}

	switch parsed.Scheme {
	case "socks5", "socks5h":
		return checkSOCKS5(parsed, host)
	case "http", "https":
		return checkTCP(host)
	default:
		return checkTCP(host)
	}
}

// checkTCP does a raw TCP dial with 5s timeout.
func checkTCP(host string) bool {
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// checkSOCKS5 validates SOCKS5 connectivity by dialing through the proxy.
func checkSOCKS5(parsed *url.URL, host string) bool {
	var auth *proxy.Auth
	if parsed.User != nil {
		pass, _ := parsed.User.Password()
		auth = &proxy.Auth{
			User:     parsed.User.Username(),
			Password: pass,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", host, auth, &net.Dialer{Timeout: 5 * time.Second})
	if err != nil {
		return false
	}

	// Try to connect through the proxy to a known host
	conn, err := dialer.Dial("tcp", "web.whatsapp.com:443")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// maskProxyURL redacts credentials from a proxy URL for safe logging.
func maskProxyURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "invalid-url"
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword("***", "***")
	}
	return parsed.String()
}
