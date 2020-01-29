package server

import (
	"github.com/rykov/paperboy/mail"

	"context"
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
	campaign, err := mail.LoadContent(r.cfg, args.Content)
	if err != nil {
		return 0, err
	}

	// Populate recipients and fire away
	campaign.Recipients = recipients
	err = mail.SendCampaign(r.cfg, campaign)
	return int32(len(recipients)), err
}
