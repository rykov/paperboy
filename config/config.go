package config

import (
	"crypto/tls"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Viper config file (populated from cmd)
var ViperConfigFile string = ""

// BuildInfo (populated from cmd)
var Build BuildInfo

type AConfig struct {
	// Version/build
	Build BuildInfo

	// From config.toml
	ConfigFile

	// Command context
	Context context.Context

	// Afero VFS
	AppFs *Fs
}

// Creates a new config with provided context override
func (orig AConfig) WithContext(ctx context.Context) *AConfig {
	var newCfg = orig
	newCfg.Context = ctx
	return &newCfg
}

// See https://www.paperboy.email/docs/configuration/
type ConfigFile struct {
	// General
	Theme string
	From  string

	// CAN-SPAM
	Address        string
	UnsubscribeURL string

	// Delivery
	SMTP   SMTPConfig
	DryRun bool

	// Validation
	DKIM map[string]interface{}

	// Directories
	AssetDir   string
	ContentDir string
	LayoutDir  string
	ThemeDir   string
	ListDir    string

	// Delivery
	SendRate float32
	Workers  int

	// Client/Server
	ClientIgnores []string
	ServerAuth    string
	ServerPort    uint
}

type SMTPConfig struct {
	URL  string
	User string
	Pass string
	TLS  *TLSConfig
}

type TLSConfig struct {
	InsecureSkipVerify bool
	MinVersion         string
}

func (t TLSConfig) GetMinVersion() (uint16, error) {
	switch t.MinVersion {
	case "":
		// Not set, so let the tls package decide
		return 0, nil
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, errors.New("Invalid TLS version")
	}
}

// Initial blank config
type BuildInfo struct {
	Version   string
	BuildDate string
}

func (i BuildInfo) String() string {
	return fmt.Sprintf("v%s %s/%s (%s)", i.Version, runtime.GOOS, runtime.GOARCH, i.BuildDate)
}

// Standard configuration with Viper operating
// on OS FS based at current working directory
func LoadConfig(ctx context.Context) (*AConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fs := afero.NewBasePathFs(afero.NewOsFs(), wd)
	return LoadConfigFs(ctx, fs)
}

// Standard configuration for specified afero FS
// The project is assumed to be in afero.Fs root
func LoadConfigFs(ctx context.Context, fs afero.Fs) (*AConfig, error) {
	cfg := &AConfig{Context: ctx, AppFs: &Fs{Fs: fs}}
	cfg.AppFs.Config = cfg

	viperConfig := newViperConfig(cfg.AppFs)
	if err := viperConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			return nil, err
		}
	}

	err := viperConfig.Unmarshal(&cfg.ConfigFile)
	return cfg, err
}

// Initialize configuration with Viper
func newViperConfig(fs afero.Fs) *viper.Viper {
	v := viper.New()

	// Initialize with real or virtual FS
	if fs != nil {
		v.SetFs(fs)
	}

	// From --config
	if ViperConfigFile != "" {
		v.SetConfigFile(ViperConfigFile)
	}

	// Tie configuration to ENV
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("paperboy")
	v.AutomaticEnv()

	// Defaults (General)
	v.SetDefault("smtp.url", "")
	v.SetDefault("smtp.user", "")
	v.SetDefault("smtp.pass", "")
	v.SetDefault("smtp.tls.InsecureSkipVerify", false)
	v.SetDefault("smtp.tls.MinVersion", "1.2")
	v.SetDefault("dryRun", false)

	// Defaults (Dirs)
	v.SetDefault("assetDir", "assets")
	v.SetDefault("contentDir", "content")
	v.SetDefault("layoutDir", "layouts")
	v.SetDefault("themeDir", "themes")
	v.SetDefault("listDir", "lists")

	// Delivery workers/rate
	v.SetDefault("sendRate", 1)
	v.SetDefault("workers", 3)

	// Server, Client, API
	v.BindEnv("serverPort", "PORT")
	v.SetDefault("serverPort", 8080)
	v.SetDefault("serverAuth", "")

	// Prepare for project's config.*
	v.SetConfigName("config")
	v.AddConfigPath("/")

	// 🐍
	return v
}
