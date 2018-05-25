package mail

import (
	"github.com/spf13/viper"

	"fmt"
	"runtime"
	"strings"
)

// Initial blank config
var Config = config{}

type config struct {
	// Version/build
	Build BuildInfo

	// From config.toml
	ConfigFile
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
	SMTP   smtpConfig
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

type smtpConfig struct {
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

// Configuration configuration :)
var viperConfig *viper.Viper

// Load configuration with Viper
func LoadConfig() error {
	viperConfig.SetFs(AppFs)
	if err := viperConfig.ReadInConfig(); err != nil {
		return err
	}
	return viperConfig.Unmarshal(&Config.ConfigFile)
}

// Initialize configuration with Viper
func InitConfig(cfgFile string) {
	viperConfig = viper.New()
	v := viperConfig

	// From --config
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
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
}
