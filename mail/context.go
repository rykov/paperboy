package mail

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// Context for campaign
type context struct {
	Recipient ctxRecipient
	Campaign  ctxCampaign

	UnsubscribeURL string
	Address        string
}

// Convert to a map (for uritemplates and debugging)
func (c *context) toFlatMap() map[string]interface{} {
	out := map[string]interface{}{}
	b, _ := json.Marshal(c)
	json.Unmarshal(b, &out)
	return flattenMap(out)
}

// Recipient variable
type ctxRecipient struct {
	Name        string
	Email       string
	Attachments []string
	Params      map[string]interface{}
}

func newRecipient(data map[string]interface{}) (*ctxRecipient, error) {
	r := &ctxRecipient{Params: keysToLower(data)}
	r.Email, _ = r.Params["email"].(string)
	r.Name, _ = r.Params["name"].(string)
	if att, ok := r.Params["attachments"].([]interface{}); ok {
		a := make([]string, len(att))
		for i, v := range att {
			attName, ok := v.(string)
			if !ok {
				continue
			}

			// Validate if the file exists or not
			p, err := filepath.Abs(AppFs.AttachmentPath(attName))
			if err != nil {
				return nil, err
			}
			if !AppFs.isFile(p) {
				return nil, fmt.Errorf("Cannot find attachment in %s", p)
			}

			a[i] = p
		}
		r.Attachments = a
	}
	delete(r.Params, "email")
	delete(r.Params, "name")
	delete(r.Params, "attachments")
	return r, nil
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
		c.From = Config.From
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
