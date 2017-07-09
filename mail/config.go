package mail

// Initial blank config
var Config = config{}

// See https://www.paperboy.email/docs/configuration/
type config struct {
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
