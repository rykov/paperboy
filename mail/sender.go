package mail

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail/send"
	"github.com/wneessen/go-mail"

	"errors"
	"fmt"
)

func LoadAndSendCampaign(cfg *config.AConfig, tmplFile, recipientFile string) error {
	// Load up template and recipientswith frontmatter
	c, err := LoadCampaign(cfg, tmplFile, recipientFile)
	if err != nil {
		return err
	}

	return SendCampaign(cfg, c)
}

func SendCampaign(cfg *config.AConfig, c *Campaign) error {
	var s send.Sender

	// Skip dial on dryRun
	if cfg.DryRun {
		s = send.NewTestSender()
	} else {
		s = send.NewSMTPSender(cfg.Context, &cfg.SMTP)
	}

	q, err := newDefaultQueue(cfg, s, c)
	if err != nil {
		return err
	}

	return sendCampaignTo(cfg, q, c)
}

func SendCampaignDryRun(cfg *config.AConfig, c *Campaign) ([][]byte, error) {
	s := send.NewTestSender()
	q, err := newDefaultQueue(cfg, s, c)
	if err != nil {
		return nil, err
	}
	err = sendCampaignTo(cfg, q, c)
	return s.Mails, err
}

func newDefaultQueue(cfg *config.AConfig, s send.Sender, c *Campaign) (send.Manager, error) {
	qc := send.Config{SendRate: cfg.SendRate, Workers: cfg.Workers, QueueID: c.ID}
	return send.NewInMemory(cfg.Context, &qc, s)
}

func sendCampaignTo(cfg *config.AConfig, queue send.Manager, c *Campaign) error {
	// Capture context cancellation for graceful exit
	done := cfg.Context.Done()
	queueErr := make(chan error, 1)

	// Async enqueue for all recipients
	go func() {
		defer close(queueErr)
		for i := range c.Recipients {
			select {
			case <-done:
				queue.Close()
				queueErr <- errors.New("stopped on context cancellation")
				return
			default:
			}

			// Render message
			m := mail.NewMsg(c.MsgOpts...)
			if err := c.renderMessage(m, i); err != nil {
				queueErr <- fmt.Errorf("could not render email for recipient %d: %w", i, err)
				queue.Close()
				return
			}

			// Enqueue message directly
			if err := queue.Enqueue(cfg.Context, m); err != nil {
				queueErr <- fmt.Errorf("failed to enqueue email: %w", err)
				queue.Close()
				return
			}
		}

		// Signal that we're done queuing
		queue.Close()
	}()

	// Wait for all tasks to complete
	errM := queue.Wait()

	// Return queueing/sending errors
	return errors.Join(errM, <-queueErr)
}
