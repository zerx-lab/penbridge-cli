// Package client implements the HTTP client and configuration for talking to
// the SiYuan kernel API (and, optionally, the read-only publish service).
package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// DefaultBaseURL is the SiYuan kernel API address. The kernel API supports
	// full read/write. The publish service (default :6808) is read-only and uses
	// Basic auth instead of a token.
	DefaultBaseURL = "http://127.0.0.1:6806"
	// DefaultTimeout is the request timeout in seconds.
	DefaultTimeout = 120
)

// Config holds the connection settings for the CLI.
type Config struct {
	// BaseURL is the kernel API base URL, e.g. http://127.0.0.1:6806.
	BaseURL string `json:"baseURL"`
	// Token is the SiYuan API token (Settings -> About -> API token). Sent as
	// "Authorization: Token <token>". Takes precedence over Basic auth.
	Token string `json:"token,omitempty"`
	// User and Password enable HTTP Basic auth (used by the publish service or a
	// reverse proxy). Only used when Token is empty.
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	// Timeout is the per-request timeout in seconds.
	Timeout int `json:"timeout,omitempty"`
	// Insecure skips TLS certificate verification (for self-signed HTTPS).
	Insecure bool `json:"insecure,omitempty"`
}

// DefaultConfig returns a Config populated with defaults.
func DefaultConfig() Config {
	return Config{BaseURL: DefaultBaseURL, Timeout: DefaultTimeout}
}

// DefaultConfigPath returns the platform-specific default config file path.
func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "penbridge", "config.json"), nil
}

// LoadConfig builds a Config from defaults, overlaying the config file (if it
// exists) and then environment variables. Missing files are not an error.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	if path != "" {
		data, err := os.ReadFile(path)
		switch {
		case err == nil:
			fileCfg := Config{}
			if err := json.Unmarshal(data, &fileCfg); err != nil {
				return cfg, fmt.Errorf("parse config %q: %w", path, err)
			}
			cfg = merge(cfg, fileCfg)
		case os.IsNotExist(err):
			// no config file yet; ignore
		default:
			return cfg, fmt.Errorf("read config %q: %w", path, err)
		}
	}
	cfg = applyEnv(cfg)
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	return cfg, nil
}

// SaveConfig writes the config to path (creating parent dirs), with 0600 perms
// since it may contain a token.
func SaveConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	return nil
}

// Redacted returns a copy of the config with secrets masked, for display.
func (c Config) Redacted() Config {
	r := c
	r.Token = mask(c.Token)
	r.Password = mask(c.Password)
	return r
}

func mask(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

// merge overlays non-zero fields of b onto a.
func merge(a, b Config) Config {
	if b.BaseURL != "" {
		a.BaseURL = b.BaseURL
	}
	if b.Token != "" {
		a.Token = b.Token
	}
	if b.User != "" {
		a.User = b.User
	}
	if b.Password != "" {
		a.Password = b.Password
	}
	if b.Timeout != 0 {
		a.Timeout = b.Timeout
	}
	if b.Insecure {
		a.Insecure = true
	}
	return a
}

func applyEnv(c Config) Config {
	if v := firstEnv("PENBRIDGE_BASE_URL", "SIYUAN_API_URL", "SIYUAN_BASE_URL"); v != "" {
		c.BaseURL = v
	}
	if v := firstEnv("PENBRIDGE_TOKEN", "SIYUAN_TOKEN", "SIYUAN_API_TOKEN"); v != "" {
		c.Token = v
	}
	if v := firstEnv("PENBRIDGE_USER", "SIYUAN_USER"); v != "" {
		c.User = v
	}
	if v := firstEnv("PENBRIDGE_PASSWORD", "SIYUAN_PASSWORD"); v != "" {
		c.Password = v
	}
	if v := firstEnv("PENBRIDGE_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Timeout = n
		}
	}
	if v := firstEnv("PENBRIDGE_INSECURE"); v != "" {
		c.Insecure = parseBool(v)
	}
	return c
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(strings.TrimSpace(s))
	return b
}
