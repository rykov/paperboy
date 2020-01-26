package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"fmt"
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

	// Afero VFS
	AppFs *Fs
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
	ContentDir string
	LayoutDir  string
	ThemeDir   string
	ListDir    string

	// Delivery
	SendRate float32
	Workers  int
}

type SMTPConfig struct {
	URL  string
	User string
	Pass string
}

// Initial blank config
type BuildInfo struct {
	Version   string
	BuildDate string
}

func (i BuildInfo) String() string {
	return fmt.Sprintf("v%s %s/%s (%s)", i.Version, runtime.GOOS, runtime.GOARCH, i.BuildDate)
}

// Initalize configuration with passed-in VFS
func NewConfig(afs afero.Fs) *AConfig {
	cfg := &AConfig{AppFs: &Fs{Fs: afs}}
	cfg.AppFs.Config = cfg
	return cfg
}

// Standard configuration with Viper
func LoadConfig() (*AConfig, error) {
	cfg := NewConfig(afero.NewOsFs()) // Config
	return cfg, LoadConfigTo(cfg)
}

// Configuration helper for tests, etc
func LoadConfigTo(cfg *AConfig) error {
	viperConfig := newViperConfig(cfg.AppFs)
	if err := viperConfig.ReadInConfig(); err != nil {
		return err
	}
	return viperConfig.Unmarshal(&cfg.ConfigFile)
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
	v.SetDefault("dryRun", false)

	// Defaults (Dirs)
	v.SetDefault("contentDir", "content")
	v.SetDefault("layoutDir", "layouts")
	v.SetDefault("themeDir", "themes")
	v.SetDefault("listDir", "lists")

	// Delivery workers/rate
	v.SetDefault("sendRate", 1)
	v.SetDefault("workers", 3)

	// Prepare for project's config.*
	v.SetConfigName("config")
	v.AddConfigPath(".")

	// 🐍
	return v
}
