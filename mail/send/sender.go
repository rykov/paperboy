package send

import (
	"crypto/tls"
	"errors"

	"github.com/cenkalti/backoff/v5"
	"github.com/wneessen/go-mail"

	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"
)

func NewSMTPSender(ctx context.Context, cfg *SMTPConfig) *smtpSender {
	return &smtpSender{ctx, cfg}
}

func NewTestSender() *testSender {
	return &testSender{}
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
		return 0, errors.New("invalid TLS version")
	}
}

type smtpSender struct {
	context context.Context
	config  *SMTPConfig
}

func (d smtpSender) NewConn() (conn Conn, err error) {
	cfg := d.config
	ctx := d.context

	// Initialize dialer from configuration
	dialer, err := smtpDialer(cfg)
	if err != nil {
		return nil, err
	}

	// Dial SMTP with 3 retries and failure logging
	return backoff.Retry(ctx, func() (*mail.Client, error) {
		err := dialer.DialWithContext(ctx)
		return dialer, err
	},
		backoff.WithMaxTries(3),
		backoff.WithBackOff(backoff.NewConstantBackOff(time.Second)),
		backoff.WithNotify(func(err error, _ time.Duration) {
			fmt.Println("Retrying SMTP dial on error: ", err)
		}),
	)
}

func smtpDialer(cfg *SMTPConfig) (*mail.Client, error) {
	surl, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	} else if surl.Host == "" {
		return nil, fmt.Errorf("invalid SMTP URL: %s", surl)
	}

	// Populate/validate scheme
	hostname := surl.Hostname()
	if s := surl.Scheme; s == "" {
		surl.Scheme = "smtps"
	} else if s != "smtp" && s != "smtps" {
		return nil, fmt.Errorf("invalid SMTP URL scheme: %s", s)
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

	// Create client options
	opts := []mail.Option{
		mail.WithPort(port),
	}

	// Add authentication if provided
	if user != "" && pass != "" {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthPlain))
		opts = append(opts, mail.WithUsername(user))
		opts = append(opts, mail.WithPassword(pass))
	}

	// Configure TLS
	if surl.Scheme == "smtps" {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	// Custom TLS config
	if cfg.TLS != nil {
		tlsMinVersion, err := cfg.TLS.GetMinVersion()
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
			MinVersion:         tlsMinVersion,
			ServerName:         hostname,
		}
		opts = append(opts, mail.WithTLSConfig(tlsConfig))
	}

	// Create the client with options
	return mail.NewClient(hostname, opts...)
}

type testSender struct {
	lock  sync.Mutex
	Mails [][]byte
}

func (s *testSender) NewConn() (conn Conn, err error) {
	return s, nil
}

func (s *testSender) Send(msgs ...*mail.Msg) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	var err error

	for _, msg := range msgs {
		var buf bytes.Buffer
		_, msgErr := msg.WriteTo(&buf)
		err = errors.Join(err, msgErr)
		if msgErr == nil {
			s.Mails = append(s.Mails, buf.Bytes())
		}
	}

	return err
}

func (s *testSender) Close() error {
	return nil
}
