package mail

import (
	"fmt"
	"runtime"
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
