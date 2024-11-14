package config

import (
	"crypto/tls"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestDefaultConfig(t *testing.T) {
	cfg := NewConfig(afero.NewMemMapFs())

	// Write and load fake configuration
	cPath, _ := filepath.Abs("./config.toml")
	afero.WriteFile(cfg.AppFs, cPath, []byte(""), 0644)
	if err := LoadConfigTo(cfg); err != nil {
		panic(err)
	}

	version, err := cfg.ConfigFile.SMTP.TLS.GetMinVersion()
	if err != nil {
		t.Error(err)
	}
	if version != tls.VersionTLS12 {
		t.Errorf("Invalid version: expected %d got %d", tls.VersionTLS12, version)
	}
	if cfg.ConfigFile.SMTP.TLS.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false")
	}
}

func TestTLSConfig(t *testing.T) {
	cfg := NewConfig(afero.NewMemMapFs())

	// Write and load fake configuration
	cPath, _ := filepath.Abs("./config.toml")
	afero.WriteFile(cfg.AppFs, cPath, []byte(`
[smtp.tls]
InsecureSkipVerify = true
MinVersion = "1.0"
	`), 0644)
	if err := LoadConfigTo(cfg); err != nil {
		panic(err)
	}

	version, err := cfg.ConfigFile.SMTP.TLS.GetMinVersion()
	if err != nil {
		t.Error(err)
	}
	if version != tls.VersionTLS10 {
		t.Errorf("Invalid version: expected %d got %d", tls.VersionTLS12, version)
	}
	if !cfg.ConfigFile.SMTP.TLS.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
}
