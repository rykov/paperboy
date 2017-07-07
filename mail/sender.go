package mail

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-gomail/gomail"
)

func SendCampaign(tmplFile, recipientFile string) error {
	// Load up template and recipientswith frontmatter
	c, err := LoadCampaign(tmplFile, recipientFile)
	if err != nil {
		return err
	}

	// Dial up the sender
	sender, err := configureSender()
	if err != nil {
		return err
	}
	defer sender.Close()

	// Send emails
	m := gomail.NewMessage()
	for i, _ := range c.Recipients {
		if err := c.renderMessage(m, i); err != nil {
			return err
		}

		fmt.Println("Sending email to ", m.GetHeader("To"))
		if err := gomail.Send(sender, m); err != nil {
			fmt.Println("  Could not send email: ", err)
		}

		// Throttle to account for quotas, etc
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

func configureSender() (sender gomail.SendCloser, err error) {
	// Dial up SMTP or dryRun
	if Config.DryRun {
		sender = &dryRunSender{}
	} else {
		dialer, err := smtpDialer(&Config.SMTP)
		if err != nil {
			return nil, err
		}
		sender, err = dialer.Dial()
		if err != nil {
			return nil, err
		}
	}

	// DKIM-signing sender, if configuration is present
	if cfg := Config.DKIM; len(cfg) > 0 {
		sender, err = SendCloserWithDKIM(sender, cfg)
		if err != nil {
			return nil, err
		}
	}

	return sender, nil
}

func smtpDialer(cfg *smtpConfig) (*gomail.Dialer, error) {
	// Dial to SMTP server (with SSL)
	surl, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	// Populate/validate scheme
	if s := surl.Scheme; s == "" {
		surl.Scheme = "smtps"
	} else if s != "smtp" && s != "smtps" {
		return nil, fmt.Errorf("Invalid SMTP URL scheme: %s", s)
	}

	// Authentication from URL
	var user, pass string
	if auth := surl.User; auth != nil {
		pass, _ = auth.Password()
		user = auth.Username()
	}

	// Authentication overrides
	if cfg.User != "" {
		user = cfg.User
	}
	if cfg.Pass != "" {
		pass = cfg.Pass
	}

	// Port
	var port int
	if i, err := strconv.Atoi(surl.Port()); err == nil {
		port = i
	} else if surl.Scheme == "smtp" {
		port = 25
	} else {
		port = 465
	}

	// Initialize the dialer
	d := gomail.NewDialer(surl.Hostname(), port, user, pass)
	d.SSL = (surl.Scheme == "smtps")
	return d, nil
}

type dryRunSender struct{}

func (s *dryRunSender) Send(from string, to []string, msg io.WriterTo) error {
	fmt.Printf("------> MAIL FROM: %s TO: %+v\n", from, to)
	msg.WriteTo(os.Stdout)
	fmt.Println("------> /MAIL")
	return nil
}

func (s *dryRunSender) Close() error {
	return nil
}
