package chaperone

import (
	"os"
	"slices"
	"strings"
	"time"

	"github.com/KillianMeersman/chaperone/pkg/config"
	"gopkg.in/yaml.v2"
)

var (
	Port               = config.GetInt64("PORT", 8080, false)
	ConfigFileLocation = config.GetString("CONFIGFILE", "./chaperone.yaml", false)
)

type RateLimit struct {
	URL          string        `yaml:"url"`
	Method       string        `yaml:"method"`
	WaitDuration time.Duration `yaml:"wait_duration"`
}

type CacheConfig struct {
	URL        string        `yaml:"url"`
	MinTTL     time.Duration `yaml:"min_ttl"`
	MaxTTL     time.Duration `yaml:"max_ttl"`
	DefaultTTL time.Duration `yaml:"default_ttl"`
}

type ConfigFile struct {
	RateLimits     []RateLimit   `yaml:"rate_limits"`
	CacheOverrides []CacheConfig `yaml:"cache_overrides"`
}

// Get the correct CacheConfig for the given url, if any exist.
// The second return value indicates if a value was found.
func (c *ConfigFile) CacheOverrideForURL(url string) (CacheConfig, bool) {
	// Copy CacheOverrides and sort based on url length.
	overrides := make([]CacheConfig, len(c.CacheOverrides))
	copy(overrides, c.CacheOverrides)
	slices.SortFunc(overrides, func(a CacheConfig, b CacheConfig) int {
		return len(b.URL) - len(a.URL)
	})

	for _, override := range overrides {
		if strings.HasPrefix(url, override.URL) {
			return override, true
		}
	}

	return CacheConfig{}, false
}

func ParseConfigFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cf := &ConfigFile{}
	err = yaml.Unmarshal(data, cf)
	if err != nil {
		return nil, err
	}

	return cf, nil
}
