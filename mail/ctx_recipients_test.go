package mail

import "testing"

func TestCtxRecipientsFilter(t *testing.T) {
	var recipients CtxRecipients = []*ctxRecipient{
		&ctxRecipient{
			Name:  "Name1",
			Email: "name1@example.com",
			Params: map[string]interface{}{
				"class": "1",
			},
		},
		&ctxRecipient{
			Name:  "Name2",
			Email: "name2@example.com",
			Params: map[string]interface{}{
				"class": "2",
			},
		},
	}
	filtered, err := recipients.Filter("class == '1'")
	if err != nil {
		t.Errorf("Failed: %s", err)
	}
	if len(filtered) != 1 {
		t.Errorf("Got %d", len(filtered))
	}
}
