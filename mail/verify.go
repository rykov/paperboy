package mail

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/rykov/paperboy/config"
)

func VerifyCampaign(cfg *config.AConfig, tmplFile, recipientFile string) error {
	// Load up template and recipients with frontmatter
	c, err := LoadCampaign(cfg, tmplFile, recipientFile)
	if err != nil {
		return fmt.Errorf("failed to load campaign: %w", err)
	}

	// Ensure dry run mode for verification
	cfg.DryRun = true
	s := &testSender{}
	if err := sendCampaignTo(s, cfg, c); err != nil {
		return fmt.Errorf("failed to render emails: %w", err)
	}

	if len(s.Mails) == 0 {
		return fmt.Errorf("no emails were rendered")
	}

	// Verify DKIM signature, if configured
	if len(cfg.ConfigFile.DKIM) != 0 {
		return verifyDKIMForMail(s.Mails[0])
	}

	// No problems
	return nil
}

// verifyDKIMForMail verifies DKIM signatures for a single email
func verifyDKIMForMail(mailData []byte) error {
	reader := bytes.NewReader(mailData)
	verifications, err := dkim.Verify(reader)
	if err != nil {
		return fmt.Errorf("DKIM verification failed: %w", err)
	}

	if len(verifications) == 0 {
		return errors.New("no DKIM signatures found")
	}

	// Collect verification errors with domain context
	var vErrs []error
	for _, v := range verifications {
		if v.Err != nil {
			vErrs = append(vErrs, fmt.Errorf("domain %s: %w", v.Domain, v.Err))
		}
	}

	if len(vErrs) > 0 {
		err := errors.Join(vErrs...)
		return fmt.Errorf("DKIM verification errors: %w", err)
	}

	return nil
}
