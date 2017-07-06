package mail

import (
	"strings"
)

// Context for campaign
type context struct {
	Recipient ctxRecipient
	Campaign  ctxCampaign
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
	From    string
	Subject string
	Params  map[string]interface{}
}

func newCampaign(data map[string]interface{}) ctxCampaign {
	c := ctxCampaign{Params: keysToLower(data)}
	c.Subject, _ = c.Params["subject"].(string)
	if c.From, _ = c.Params["from"].(string); c.From == "" {
		c.From = Config.GetString("from")
	}

	delete(c.Params, "subject")
	delete(c.Params, "from")
	return c
}

func keysToLower(data map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range data {
		out[strings.ToLower(k)] = v
	}
	return out
}
