package mail

import (
	"fmt"
	"strings"
	"testing"
)

func TestSmtpDialerSuccess(t *testing.T) {
	cases := []struct {
		testName string

		// Inputs
		smtpURL  string
		smtpUser string
		smtpPass string

		// Expected outputs
		host string
		user string
		pass string
		port int
		ssl  bool
	}{
		{"Full configuration from URL",
			"smtps://hello:world@smtp.host:1199", "", "",
			"smtp.host", "hello", "world", 1199, true,
		},
		{"Defaults for everything",
			"//smtp.host", "", "",
			"smtp.host", "", "", 465, true,
		},
		{`Defaults for "smtps"`,
			"smtps://smtp.host", "", "",
			"smtp.host", "", "", 465, true,
		},
		{`Defaults for "smtp"`,
			"smtp://smtp.host", "", "",
			"smtp.host", "", "", 25, false,
		},
		{"Username override",
			"//hello:world@smtp.host", "bye", "",
			"smtp.host", "bye", "world", 465, true,
		},
		{"Password override",
			"//hello:world@smtp.host", "", "earth",
			"smtp.host", "hello", "earth", 465, true,
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf(c.testName), func(t *testing.T) {
			d, err := smtpDialer(&smtpConfig{
				URL:  c.smtpURL,
				User: c.smtpUser,
				Pass: c.smtpPass,
			})

			if err != nil {
				t.Errorf("Dialer initialization error: %s ", err)
			} else if d.SSL != c.ssl {
				t.Errorf("Dialer incorrect SSL: %t", d.SSL)
			} else if d.Host != c.host {
				t.Errorf("Dialer has invalid host: %s", d.Host)
			} else if d.Port != c.port {
				t.Errorf("Dialer has invalid post: %d", d.Port)
			} else if d.Username != c.user {
				t.Errorf("Dialer has invalid user: %s", d.Username)
			} else if d.Password != c.pass {
				t.Errorf("Dialer has invalid pass: %s", d.Password)
			}
		})
	}
}

func TestSmtpDialerFailure(t *testing.T) {
	cases := []struct {
		smtpURL string
		err     string
	}{
		{"%gh&%ij", "invalid URL"},
		{"https://host", "Invalid SMTP URL scheme"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf(c.err), func(t *testing.T) {
			_, err := smtpDialer(&smtpConfig{URL: c.smtpURL})
			if err == nil {
				t.Errorf("Dialer should cause an error")
			} else if !strings.Contains(err.Error(), c.err) {
				t.Errorf("Dialer error %q should contain %q", err, c.err)
			}
		})
	}
}
