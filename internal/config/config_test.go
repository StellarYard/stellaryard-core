package config

import (
	"os"
	"testing"
)

func TestConfigDefaultsToLoopback(t *testing.T) {
	os.Unsetenv("STELLARYARD_HOST")
	os.Unsetenv("STELLARYARD_PORT")
	os.Unsetenv("STELLARYARD_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want 127.0.0.1", cfg.Host)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if !cfg.IsLoopback() {
		t.Error("IsLoopback() should be true for 127.0.0.1")
	}
	if cfg.Address() != "127.0.0.1:8080" {
		t.Errorf("Address() = %q, want 127.0.0.1:8080", cfg.Address())
	}
}

func TestConfigNonLoopbackKeyScenarios(t *testing.T) {
	const valid32Key = "12345678901234567890123456789012" // exactly 32 chars
	const validLongKey = "a-very-secure-random-token-exceeding-thirty-two-bytes-123456789"

	tests := []struct {
		name      string
		host      string
		key       string
		setKey    bool
		wantError bool
	}{
		// Non-loopback tests
		{
			name:      "non-loopback missing key fails",
			host:      "0.0.0.0",
			setKey:    false,
			wantError: true,
		},
		{
			name:      "non-loopback empty key fails",
			host:      "0.0.0.0",
			key:       "",
			setKey:    true,
			wantError: true,
		},
		{
			name:      "non-loopback whitespace-only key fails",
			host:      "0.0.0.0",
			key:       "    \t\n  ",
			setKey:    true,
			wantError: true,
		},
		{
			name:      "non-loopback short key (<32 chars) fails",
			host:      "0.0.0.0",
			key:       "short-secret-key-12345",
			setKey:    true,
			wantError: true,
		},
		{
			name:      "non-loopback valid 32-char key succeeds",
			host:      "0.0.0.0",
			key:       valid32Key,
			setKey:    true,
			wantError: false,
		},
		{
			name:      "non-loopback valid long key succeeds",
			host:      "0.0.0.0",
			key:       validLongKey,
			setKey:    true,
			wantError: false,
		},
		{
			name:      "non-loopback custom IP with valid key succeeds",
			host:      "192.168.1.100",
			key:       valid32Key,
			setKey:    true,
			wantError: false,
		},

		// Loopback tests (preserves unauthenticated dev mode)
		{
			name:      "loopback 127.0.0.1 missing key succeeds",
			host:      "127.0.0.1",
			setKey:    false,
			wantError: false,
		},
		{
			name:      "loopback localhost missing key succeeds",
			host:      "localhost",
			setKey:    false,
			wantError: false,
		},
		{
			name:      "loopback ::1 missing key succeeds",
			host:      "::1",
			setKey:    false,
			wantError: false,
		},
		{
			name:      "loopback 127.0.0.1 with short key succeeds (local dev)",
			host:      "127.0.0.1",
			key:       "short-key",
			setKey:    true,
			wantError: false,
		},
		{
			name:      "loopback 127.0.0.1 with strong key succeeds",
			host:      "127.0.0.1",
			key:       valid32Key,
			setKey:    true,
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("STELLARYARD_HOST", tc.host)
			defer os.Unsetenv("STELLARYARD_HOST")

			if tc.setKey {
				os.Setenv("STELLARYARD_API_KEY", tc.key)
				defer os.Unsetenv("STELLARYARD_API_KEY")
			} else {
				os.Unsetenv("STELLARYARD_API_KEY")
			}

			cfg, err := Load()
			if tc.wantError {
				if err == nil {
					t.Errorf("Load() expected error for host=%q key=%q, got nil", tc.host, tc.key)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error for host=%q key=%q: %v", tc.host, tc.key, err)
				}
				if cfg != nil && tc.setKey && cfg.APIKey != tc.key {
					t.Errorf("Load() preserved key = %q, want %q (should not transform secret)", cfg.APIKey, tc.key)
				}
			}
		})
	}
}
