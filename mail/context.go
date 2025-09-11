package mail

import (
	"github.com/rykov/paperboy/config"
	log "github.com/sirupsen/logrus"

	"encoding/json"
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
type ctxRecipient struct {
	Name   string
	Email  string
	Params map[string]interface{}
}

func newRecipient(data map[string]interface{}) ctxRecipient {
	r := ctxRecipient{Params: keysToLower(data)}
	r.Email, _ = r.Params["email"].(string)
	r.Name, _ = r.Params["name"].(string)
	delete(r.Params, "email")
	delete(r.Params, "name")
	return r
}

// Campaign variable
type ctxCampaign struct {
	From string
	// Filter recipients
	Filter string
	Params map[string]interface{}

	// Original subject from frontmatter
	// before templating via renderSubject
	subject string
}

func (c ctxCampaign) Subject() string {
	log.Warnf("{{ .Campaign.Subject }} is deprecated, use {{ .Subject }}")
	return c.subject
}

func newCampaign(cfg *config.AConfig, data map[string]interface{}) ctxCampaign {
	c := ctxCampaign{Params: keysToLower(data)}
	c.subject, _ = c.Params["subject"].(string)
	if c.From, _ = c.Params["from"].(string); c.From == "" {
		c.From = cfg.From
	}
	c.Filter, _ = c.Params["filter"].(string)

	delete(c.Params, "subject")
	delete(c.Params, "from")
	delete(c.Params, "filter")
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
