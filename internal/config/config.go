package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Config holds runtime server configuration.
type Config struct {
	Host   string
	Port   string
	APIKey string
}

// Load reads and validates server configuration from the environment.
func Load() (*Config, error) {
	host := os.Getenv("STELLARYARD_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("STELLARYARD_PORT")
	if port == "" {
		port = "8080"
	}

	apiKey := os.Getenv("STELLARYARD_API_KEY")

	cfg := &Config{
		Host:   host,
		Port:   port,
		APIKey: apiKey,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// IsLoopback reports whether the configured Host is a loopback address.
func (c *Config) IsLoopback() bool {
	h := strings.TrimSpace(c.Host)
	if h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return true
	}
	ip := net.ParseIP(h)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}

// Address returns the combined host:port address for net.Listen.
func (c *Config) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

// Validate checks safety constraints: non-loopback listeners MUST have a strong STELLARYARD_API_KEY (>= 32 chars).
func (c *Config) Validate() error {
	if !c.IsLoopback() {
		trimmed := strings.TrimSpace(c.APIKey)
		if trimmed == "" {
			return fmt.Errorf("refusing to bind to non-loopback host %q without STELLARYARD_API_KEY configured", c.Host)
		}
		if len(c.APIKey) < 32 {
			return fmt.Errorf("STELLARYARD_API_KEY must be at least 32 characters for non-loopback host %q (got %d chars)", c.Host, len(c.APIKey))
		}
	}
	return nil
}
