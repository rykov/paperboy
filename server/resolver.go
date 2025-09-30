package server

import (
	"github.com/jordan-wright/email"
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail"

	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ===== ROOT QUERY RESOLVER ======

type Resolver struct {
	cfg *config.AConfig
}

func (r *Resolver) RenderOne(ctx context.Context, args *RenderOneArgs) (*renderedEmail, error) {
	i := strings.LastIndex(args.Recipient, "#")
	if i < 0 {
		return nil, fmt.Errorf("please specify one recipient with \"#\"")
	}

	listID, recIDstr := args.Recipient[0:i], args.Recipient[i+1:]
	recID, err := strconv.Atoi(recIDstr)
	if err != nil {
		return nil, fmt.Errorf("specifier should be a number: %s", recIDstr)
	}

	// Request config with context
	cfg := r.cfg.WithContext(ctx)

	// Load campaign and recipient list
	campaign, err := mail.LoadCampaign(cfg, args.Content, listID)
	if err != nil {
		return nil, err
	} else if len(campaign.Recipients) == 0 {
		return nil, fmt.Errorf("no recipients in list %s", listID)
	}

	// Find message for specified recipient
	msg, err := campaign.MessageFor(recID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := msg.WriteTo(&buf); err != nil {
		return nil, err
	}

	out := &renderedEmail{raw: buf.String()}
	if out.msg, err = email.NewEmailFromReader(&buf); err != nil {
		return nil, err
	}

	return out, nil
}

type RenderOneArgs struct {
	Content   string
	Recipient string
}

// ===== Rendered Email TYPE ======
type renderedEmail struct {
	raw string
	msg *email.Email
}

func (e *renderedEmail) RawMessage() string {
	return e.raw
}

func (e *renderedEmail) Text() string {
	return string(e.msg.Text)
}

func (e *renderedEmail) HTML() *string {
	if len(e.msg.HTML) == 0 {
		return nil
	}
	out := string(e.msg.HTML)
	return &out
}

// ===== Build/Version information =====

func (r *Resolver) PaperboyInfo(ctx context.Context) *paperboyInfo {
	return &paperboyInfo{r.cfg.Build}
}

type paperboyInfo struct {
	b config.BuildInfo
}

func (i *paperboyInfo) Version() string {
	return i.b.Version
}

func (i *paperboyInfo) BuildDate() string {
	return i.b.BuildDate
}
