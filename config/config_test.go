package config

import (
	"crypto/tls"
	"testing"

	"github.com/spf13/afero"
)

func TestDefaultConfig(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Write and load fake configuration
	afero.WriteFile(fs, "/config.toml", []byte(""), 0644)
	cfg, err := LoadConfigFs(t.Context(), fs)
	if err != nil {
		t.Fatal(err)
	}

	version, err := cfg.SMTP.TLS.GetMinVersion()
	if err != nil {
		t.Error(err)
	}
	if version != tls.VersionTLS12 {
		t.Errorf("Invalid version: expected %d got %d", tls.VersionTLS12, version)
	}
	if cfg.SMTP.TLS.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false")
	}
}

func TestTLSConfig(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Write and load fake configuration
	afero.WriteFile(fs, "/config.toml", []byte(`
[smtp.tls]
InsecureSkipVerify = true
MinVersion = "1.0"
	`), 0644)
	cfg, err := LoadConfigFs(t.Context(), fs)
	if err != nil {
		t.Fatal(err)
	}

	version, err := cfg.SMTP.TLS.GetMinVersion()
	if err != nil {
		t.Error(err)
	}
	if version != tls.VersionTLS10 {
		t.Errorf("Invalid version: expected %d got %d", tls.VersionTLS12, version)
	}
	if !cfg.SMTP.TLS.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
}
