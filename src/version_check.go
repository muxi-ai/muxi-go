package muxi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	sdkName     = "go"
	cacheFile   = ".muxi/sdk-versions.json"
	twelveHours = 12 * time.Hour
)

var (
	versionCheckOnce sync.Once
)

type versionEntry struct {
	Current      string `json:"current"`
	Latest       string `json:"latest"`
	LastNotified string `json:"last_notified,omitempty"`
}

type versionCache map[string]*versionEntry

func isDevMode() bool {
	return os.Getenv("MUXI_DEBUG") != ""
}

func getCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, cacheFile)
}

func loadCache() versionCache {
	path := getCachePath()
	if path == "" {
		return make(versionCache)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return make(versionCache)
	}

	var cache versionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return make(versionCache)
	}
	return cache
}

func saveCache(cache versionCache) {
	path := getCachePath()
	if path == "" {
		return
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
}

func wasNotifiedRecently() bool {
	cache := loadCache()
	entry := cache[sdkName]
	if entry == nil || entry.LastNotified == "" {
		return false
	}

	lastNotified, err := time.Parse(time.RFC3339, entry.LastNotified)
	if err != nil {
		return false
	}
	return time.Since(lastNotified) < twelveHours
}

func updateLatestVersion(latest string) {
	cache := loadCache()
	entry := cache[sdkName]
	if entry == nil {
		entry = &versionEntry{}
	}
	entry.Current = Version
	entry.Latest = latest
	cache[sdkName] = entry
	saveCache(cache)
}

func markNotified() {
	cache := loadCache()
	if entry := cache[sdkName]; entry != nil {
		entry.LastNotified = time.Now().Format(time.RFC3339)
		saveCache(cache)
	}
}

func isNewerVersion(latest, current string) bool {
	return latest > current
}

// CheckForUpdates checks response headers for SDK update notification.
// Called once per process, dev mode only.
func CheckForUpdates(resp *http.Response) {
	versionCheckOnce.Do(func() {
		if !isDevMode() {
			return
		}
		if resp == nil {
			return
		}

		latest := resp.Header.Get("X-Muxi-SDK-Latest")
		if latest == "" {
			return // Old server, no header
		}

		if !isNewerVersion(latest, Version) {
			return
		}

		updateLatestVersion(latest)

		if !wasNotifiedRecently() {
			println("[muxi] SDK update available:", latest, "(current:", Version+")")
			println("[muxi] Run: go get -u github.com/muxi-ai/muxi-go@latest")
			markNotified()
		}
	})
}
