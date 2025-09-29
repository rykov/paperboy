package mail

import (
	"github.com/rykov/paperboy/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"encoding/json"
	"slices"
	"strings"
)

// Shared variables for URL and template rendering
type renderContext struct {
	Recipient ctxRecipient
	Campaign  ctxCampaign

	UnsubscribeURL string
	Address        string
}

// Convert to a map (for uritemplates and debugging)
func (c *renderContext) toFlatMap() map[string]interface{} {
	out := map[string]interface{}{}
	b, _ := json.Marshal(c)
	json.Unmarshal(b, &out)
	return flattenMap(out)
}

// Recipient variable
type ctxRecipient map[string]any

func (r ctxRecipient) Name() string {
	return cast.ToString(r["name"])
}

func (r ctxRecipient) Email() string {
	return cast.ToString(r["email"])
}

func newRecipient(data map[string]any) ctxRecipient {
	return ctxRecipient(keysToLower(data))
}

// Campaign variable
type ctxCampaign struct {
	From   string
	Params map[string]interface{}

	// Original subject from frontmatter
	// before templating via renderSubject
	subject string

	// Original "To" from frontmatter
	// before templating via addMessageRecipient
	to string

	// Paths to attachments to each email
	attachments []string
}

func (c ctxCampaign) Subject() string {
	log.Warnf("{{ .Campaign.Subject }} is deprecated, use {{ .Subject }}")
	return c.subject
}

func newCampaign(cfg *config.AConfig, data map[string]interface{}) ctxCampaign {
	c := ctxCampaign{Params: keysToLower(data)}
	c.subject = cast.ToString(c.Params["subject"])
	c.to = cast.ToString(c.Params["to"])

	c.From = cast.ToString(c.Params["from"])
	if c.From == "" {
		c.From = cfg.From
	}

	// This will cast either an array or an invidivual string into an array.
	// We remove blanks because an empty string will become []string{""}
	if ary, err := cast.ToStringSliceE(c.Params["attachments"]); err == nil {
		c.attachments = slices.DeleteFunc(ary, func(s string) bool {
			return s == "" // delete blanks
		})
	}

	delete(c.Params, "attachments")
	delete(c.Params, "subject")
	delete(c.Params, "from")
	delete(c.Params, "to")
	return c
}

func keysToLower(data map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range data {
		out[strings.ToLower(k)] = v
	}
	return out
}

// Takes nested maps and brings all keys to top level with dot separators
// We use this to pass context variables to "uritemplate" library
func flattenMap(input map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range input {
		if m, ok := v.(map[string]interface{}); ok {
			for i, j := range flattenMap(m) {
				out[k+"."+i] = j
			}
		} else {
			out[k] = v
		}
	}
	return out
}
