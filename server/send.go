package server

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero/zipfs"

	"archive/zip"
	"context"
	"errors"
	"fmt"
)

type SendOneArgs struct {
	Recipients []RecipientArg
	Content    string
}

type RecipientArg struct {
	Email  string
	Name   *string
	Params *map[string]interface{}
}

// ===== Deliver campaign to one or more recipients ======
func (r *Resolver) SendBeta(ctx context.Context, args SendOneArgs) (int32, error) {
	// Some reflect voodoo is happening here with the nested array
	if len(args.Recipients) == 0 {
		return 0, fmt.Errorf("No recipients")
	}

	// Request config with context
	cfg := r.cfg.WithContext(ctx)

	// Prepare argument for mail.MapsToRecipients
	paramsAry := make([]map[string]interface{}, len(args.Recipients))
	for i, r := range args.Recipients {
		if r.Params == nil {
			paramsAry[i] = map[string]interface{}{}
		} else {
			paramsAry[i] = *r.Params
		}
	}

	// Marshal recipients into an array of ctxRecipient's
	recipients, err := mail.MapsToRecipients(paramsAry)
	if err != nil {
		return 0, err
	} else if len(recipients) == 0 {
		return 0, fmt.Errorf("No recipients")
	}

	// Validate all recipients
	for i, r := range recipients {
		r.Email = args.Recipients[i].Email
		if r.Email == "" {
			return 0, fmt.Errorf("No email for recipient #%d", i)
		}
	}

	// Load content and metadata
	campaign, err := mail.LoadContent(cfg, args.Content)
	if err != nil {
		return 0, err
	}

	// Populate recipients and fire away
	campaign.Recipients = recipients
	err = mail.SendCampaign(cfg, campaign)
	return int32(len(recipients)), err
}

type SendCampaignArgs struct {
	Campaign string
	List     string
}

// ===== Use ZIP-file attachment to deliver campaign to the recipient list ======
func (r *Resolver) SendCampaign(ctx context.Context, args SendCampaignArgs) (bool, error) {
	file, ok := RequestZipFile(ctx)
	if !ok {
		return false, errors.New("ZIP: No file")
	}
	fi, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("ZIP: %w", err)
	}
	zr, err := zip.NewReader(file, fi.Size())
	if err != nil {
		return false, fmt.Errorf("ZIP: %w", err)
	}

	// Wrap incoming ZIP file into a virtual FS with context
	cfg, err := config.LoadConfigFs(ctx, zipfs.New(zr))
	if err != nil {
		return false, fmt.Errorf("ZIP Config: %w", err)
	}

	// Load campaign and recipient list, and send it 🚀
	err = mail.LoadAndSendCampaign(cfg, args.Campaign, args.List)
	return err == nil, err
}
