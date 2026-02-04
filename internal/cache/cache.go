package cache

import (
	"fmt"
	"os/exec"
)

// Manager handles all caching operations
type Manager struct{}

// New creates a new cache manager
func New() *Manager {
	return &Manager{}
}

// Stats represents cache statistics
type Stats struct {
	VarnishHitRate    float64
	DragonflyHitRate  float64
	VarnishHits       int64
	VarnishMisses     int64
	DragonflyMemory   string
	DragonflyKeys     int64
}

// GetStats retrieves cache statistics
func (m *Manager) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Get Varnish stats
	if out, err := exec.Command("varnishstat", "-1", "-f", "MAIN.cache_hit", "-f", "MAIN.cache_miss").Output(); err == nil {
		fmt.Sscanf(string(out), "MAIN.cache_hit %d\nMAIN.cache_miss %d", &stats.VarnishHits, &stats.VarnishMisses)
		if stats.VarnishHits+stats.VarnishMisses > 0 {
			stats.VarnishHitRate = float64(stats.VarnishHits) / float64(stats.VarnishHits+stats.VarnishMisses) * 100
		}
	}

	// Get DragonflyDB stats
	if out, err := exec.Command("redis-cli", "INFO", "memory").Output(); err == nil {
		fmt.Sscanf(string(out), "used_memory_human:%s", &stats.DragonflyMemory)
	}

	if out, err := exec.Command("redis-cli", "DBSIZE").Output(); err == nil {
		fmt.Sscanf(string(out), "(integer) %d", &stats.DragonflyKeys)
	}

	return stats, nil
}

// PurgeVarnish purges Varnish cache
func (m *Manager) PurgeVarnish(pattern string) error {
	if pattern == "" {
		pattern = "."
	}
	return exec.Command("varnishadm", "ban", fmt.Sprintf("req.url ~ %s", pattern)).Run()
}

// PurgeVarnishAll purges entire Varnish cache
func (m *Manager) PurgeVarnishAll() error {
	return m.PurgeVarnish(".")
}

// PurgeVarnishURL purges a specific URL
func (m *Manager) PurgeVarnishURL(url string) error {
	return exec.Command("varnishadm", "ban", fmt.Sprintf("req.url == %s", url)).Run()
}

// FlushDragonfly flushes DragonflyDB
func (m *Manager) FlushDragonfly() error {
	return exec.Command("redis-cli", "FLUSHALL").Run()
}

// FlushDragonflyDB flushes a specific database
func (m *Manager) FlushDragonflyDB(db int) error {
	return exec.Command("redis-cli", "-n", fmt.Sprintf("%d", db), "FLUSHDB").Run()
}

// PurgeOPCache purges PHP OPCache via WP-CLI
func (m *Manager) PurgeOPCache(sitePath string) error {
	return exec.Command("wp", "eval", "opcache_reset();", "--path="+sitePath+"/public").Run()
}

// PurgeAll purges all caches
func (m *Manager) PurgeAll(sitePath string) error {
	m.PurgeVarnishAll()
	m.FlushDragonfly()
	m.PurgeOPCache(sitePath)
	return nil
}

// WarmCache warms the cache by visiting key pages
func (m *Manager) WarmCache(urls []string) error {
	for _, url := range urls {
		exec.Command("curl", "-s", "-o", "/dev/null", url).Run()
	}
	return nil
}

// VarnishStatus returns Varnish service status
func (m *Manager) VarnishStatus() (bool, error) {
	err := exec.Command("systemctl", "is-active", "--quiet", "varnish").Run()
	return err == nil, nil
}

// DragonflyStatus returns DragonflyDB status
func (m *Manager) DragonflyStatus() (bool, error) {
	out, err := exec.Command("docker", "ps", "--filter", "name=dragonfly", "-q").Output()
	return len(out) > 0, err
}
