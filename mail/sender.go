package mail

import (
	"fmt"
	"io"
	"net/url"
	"os"
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
	// Dial up the sender or dryRun
	if Config.GetBool("dryRun") {
		sender = &dryRunSender{}
	} else {
		sender, err = dialSMTPURL(Config.GetString("smtp.url"))
		if err != nil {
			return nil, err
		}
	}

	// DKIM-signing sender, if configuration is present
	if cfg := Config.GetStringMap("dkim"); len(cfg) > 0 {
		sender, err = SendCloserWithDKIM(sender, cfg)
		if err != nil {
			return nil, err
		}
	}

	return sender, nil
}

func dialSMTPURL(smtpURL string) (gomail.SendCloser, error) {
	// Dial to SMTP server (with SSL)
	surl, err := url.Parse(smtpURL)
	if err != nil {
		return nil, err
	}

	// Authentication
	user, pass := Config.GetString("smtp.user"), Config.GetString("smtp.pass")
	if auth := surl.User; auth != nil {
		pass, _ = auth.Password()
		user = auth.Username()
	}

	// TODO: Split & parse port from url.Host
	host, port := surl.Host, 465

	// Dial SMTP server
	d := gomail.NewDialer(host, port, user, pass)
	d.SSL = true // Force SSL (TODO: use schema)
	return d.Dial()
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
