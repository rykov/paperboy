package mail

import (
	"crypto/tls"
	"errors"

	"github.com/cenkalti/backoff/v5"
	"github.com/rykov/paperboy/config"
	"github.com/wneessen/go-mail"

	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// SendCloser interface for sending emails
type SendCloser interface {
	Send(msg ...*mail.Msg) error
	Close() error
}

func LoadAndSendCampaign(cfg *config.AConfig, tmplFile, recipientFile string) error {
	// Load up template and recipientswith frontmatter
	c, err := LoadCampaign(cfg, tmplFile, recipientFile)
	if err != nil {
		return err
	}

	return SendCampaign(cfg, c)
}

func SendCampaign(cfg *config.AConfig, c *Campaign) error {
	// Initialize deliverer
	engine := &deliverer{
		tasks:    make(chan *mail.Msg, 10),
		waiter:   &sync.WaitGroup{},
		context:  cfg.Context,
		campaign: c,
	}

	// Capture context cancellation for graceful exit
	done := cfg.Context.Done()
	go func() {
		<-done
		engine.close()
	}()

	// Rate configuration
	throttle, workers := time.Duration(0), cfg.Workers
	if cfg.SendRate > 0 {
		throttle = time.Duration(1000 / cfg.SendRate)
		throttle = throttle * time.Millisecond
	}

	// Start queueing emails to keep workers from idling
	fmt.Printf("Sending an email every %s via %d workers\n", throttle, workers)
	go func() {
		for i := range c.Recipients {
			m := mail.NewMsg(c.MsgOpts...)
			if err := c.renderMessage(m, i); err != nil {
				fmt.Printf("Could not queue email: %s\n", err)
				engine.close()
				return
			}

			// Gracefully handle exits and make sure we don't
			// try to queue mails into a closed channel
			if engine.stop {
				fmt.Printf("Stopped queing before %+v\n", m.GetToString())
				break
			} else {
				engine.tasks <- m
				if throttle > 0 {
					time.Sleep(throttle)
				}
			}
		}
		engine.close()
	}()

	// Start delivery workers
	for i := 0; i < workers; i++ {
		if err := engine.startWorker(i); err != nil {
			engine.close()
			return err
		}

		// HACK: Avoid race warning
		engine.stopL.Lock()
		stopped := engine.stop
		engine.stopL.Unlock()
		if stopped {
			break
		}
	}

	// Wait until everything is done
	engine.waiter.Wait()
	return nil
}

type deliverer struct {
	campaign *Campaign
	context  context.Context
	waiter   *sync.WaitGroup
	tasks    chan *mail.Msg

	stop  bool
	stopL sync.Mutex

	// Go-Mail middleware
	middleware []mail.Middleware
}

func (d *deliverer) close() {
	d.stopL.Lock()
	defer d.stopL.Unlock()

	if !d.stop {
		d.stop = true
		close(d.tasks)
	}
}

// Note: gomail doesn't expose a connection reset method,
// so the only way to clear an errored connection is to
// disconnect, and start over. Maybe we should explore
// using an alternative library.

func (d *deliverer) startWorker(id int) error {
	fmt.Printf("[%d] Starting worker...\n", id)
	d.waiter.Add(1)

	// Dial up the sender
	sender, err := d.configureSender()
	if err != nil {
		return err
	}

	go func() {
		defer d.waiter.Done()
		defer fmt.Printf("[%d] Stopping worker...\n", id)
		defer sender.Close()
		c := d.campaign

		for {
			m, more := <-d.tasks
			if !more {
				return
			}
			fmt.Printf("[%d] Sending %s to %s\n", id, c.ID, m.GetToString())
			if err := sender.Send(m); err != nil {
				fmt.Printf("[%d] Could not send email: %s\n", id, err)
				sender.Close() // Replace errored connection
				sender, err = d.configureSender()
				if err != nil {
					break
				}
			}
		}
	}()

	return nil
}

func (d *deliverer) configureSender() (sender SendCloser, err error) {
	cfg := d.campaign.Config
	ctx := d.context // not used with go-mail

	// Skip dial on dryRun
	if cfg.DryRun {
		return &dryRunSender{Mails: [][]byte{}}, nil
	}

	// Initialize dialer from configuration
	dialer, err := smtpDialer(&cfg.SMTP)
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

func smtpDialer(cfg *config.SMTPConfig) (*mail.Client, error) {
	surl, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	} else if surl.Host == "" {
		return nil, fmt.Errorf("Invalid SMTP URL: %s", surl)
	}

	// Populate/validate scheme
	hostname := surl.Hostname()
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

// Allow testing deliveries in libraries via dryRun
var LastRunResult *dryRunSender

type dryRunSender struct {
	lock  sync.Mutex
	Mails [][]byte
}

func (s *dryRunSender) Send(msgs ...*mail.Msg) error {
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

func (s *dryRunSender) Close() error {
	LastRunResult = s
	return nil
}
