package mail

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail/send"
)

func VerifyCampaign(cfg *config.AConfig, tmplFile, recipientFile string) error {
	// Load up template and recipients with frontmatter
	c, err := LoadCampaign(cfg, tmplFile, recipientFile)
	if err != nil {
		return fmt.Errorf("failed to load campaign: %w", err)
	}

	// Check for duplicate recipient email addresses
	if err := checkDuplicateEmails(c.Recipients); err != nil {
		return err
	}

	// Validate recipient parameters against schema if schema exists
	if err := verifyRecipientSchema(cfg.AppFs, tmplFile, c.EmailMeta, c.Recipients); err != nil {
		return err
	}

	// Ensure dry run mode for verification
	cfg.DryRun = true
	s := send.NewTestSender()
	q, err := newDefaultQueue(cfg, s, c)
	if err != nil {
		return fmt.Errorf("failed to start queue: %w", err)
	}

	if err := sendCampaignTo(cfg, q, c); err != nil {
		return fmt.Errorf("failed to render emails: %w", err)
	}

	if len(s.Mails) == 0 {
		return fmt.Errorf("no emails were rendered")
	}

	// Verify DKIM signature, if configured
	if len(cfg.DKIM) != 0 {
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

// checkDuplicateEmails verifies that there are no duplicate email addresses in the recipient list
func checkDuplicateEmails(recipients []*ctxRecipient) error {
	seen := make(map[string]int)

	for i, recipient := range recipients {
		// Normalize email address: trim whitespace and convert to lowercase
		email := strings.TrimSpace(strings.ToLower(recipient.Email()))

		if email == "" {
			return fmt.Errorf("recipient at index %d has empty email address", i)
		}

		if firstIndex, exists := seen[email]; exists {
			return fmt.Errorf("duplicate email address found: %q (first seen at index %d, duplicate at index %d)",
				recipient.Email(), firstIndex, i)
		}

		seen[email] = i
	}

	return nil
}

// verifyRecipientSchema validates recipients against their schema if it exists
func verifyRecipientSchema(appFs *config.Fs, tmplFile string, campaign *ctxCampaign, recipients []*ctxRecipient) error {
	// Extract template ID from file path
	tmplID := strings.TrimSuffix(tmplFile, filepath.Ext(tmplFile))

	// Load schema (custom if exists, default if no custom "to" template, or none)
	schema, err := loadRecipientSchemaWithDefault(appFs, tmplID, campaign)
	if err != nil {
		return fmt.Errorf("failed to load recipient schema: %w", err)
	}

	// If no schema exists, skip validation
	if schema == nil {
		return nil
	}

	schemaName := strings.TrimSuffix(schema.Location, "#")
	schemaName = strings.TrimPrefix(schemaName, "schema://")
	fmt.Printf("Validating recipients with %s\n", schemaName)

	// Validate each recipient against the schema
	for i, recipient := range recipients {
		// Convert to regular map for JSON schema validation
		validationData := map[string]any(*recipient)
		if err := schema.Validate(validationData); err != nil {
			return fmt.Errorf("recipient schema validation failed: recipient %d: %w", i, err)
		}
	}

	return nil
}
