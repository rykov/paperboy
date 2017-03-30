package mail

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/go-gomail/gomail"
)

func configureSender() (sender gomail.SendCloser, err error) {
	// Dial up the sender or dryRun
	if Config.GetBool("dryRun") {
		sender = &dryRunSender{}
	} else {
		sender, err = dialSMTPURL(Config.GetString("smtpURL"))
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
	user, pass := Config.GetString("smtpUser"), Config.GetString("smtpPass")
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
