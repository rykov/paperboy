package server

import (
	"github.com/google/go-cmp/cmp"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"

	"bytes"
	"encoding/json"
	netmail "net/mail"
	"testing"
)

func TestSendMutation(t *testing.T) {
	cfg, fs := newTestConfigAndFs(t)
	afero.WriteFile(fs, fs.ContentPath("c1.md"), []byte("# Hello"), 0644)
	afero.WriteFile(fs, fs.ContentPath("sub/c2.md"), []byte("# World"), 0644)
	afero.WriteFile(fs, fs.ContentPath("skip.txt"), []byte("Not-content"), 0644)

	recipients := []interface{}{
		map[string]interface{}{
			"email":  "test1@example.com",
			"params": map[string]interface{}{"name": "Test1"},
		},
		map[string]interface{}{
			"email":  "test2@example.com",
			"params": map[string]interface{}{"name": "Test2"},
		},
	}

	response := issueGraphQL(cfg, `
    mutation send($content: String!, $recipients: [RecipientInput!]!) {
      sendBeta(content: $content, recipients: $recipients)
    }
  `, map[string]interface{}{
		"recipients": recipients,
		"content":    "c1",
	})

	// VERIFY GRAPHQL RESPONSE

	if errs := response.Errors; len(errs) > 0 {
		t.Fatalf("GraphQL errors %+v", errs)
	}

	data := struct{ SendBeta int }{}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("GraphQL data JSON error: %s", err)
	}

	if a, e := data.SendBeta, len(recipients); a != e {
		t.Fatalf("GraphQL sendBeta expected %d, got %d", e, a)
	}

	// VERIFY TEST DELIVERIES

	lastRun := mail.LastRunResult
	if lastRun == nil {
		t.Fatalf("Missing dry run results")
	}

	if a, e := len(lastRun.Mails), len(recipients); a != e {
		t.Fatalf("Number of mails should be %d, got %d", e, a)
	}

	// Compare recipient metadata for each mail
	actualMeta := []interface{}{}
	for i, raw := range lastRun.Mails {
		m, err := netmail.ReadMessage(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("Error parsing delivery #%d: %s", i, err)
		}

		toList, err := m.Header.AddressList("To")
		if err != nil {
			t.Fatalf("Error parsing recipients #%d: %s", i, err)
		} else if len(toList) != 1 {
			t.Fatalf("Non-single recipient #%d", i)
		}

		actualMeta = append(actualMeta, map[string]interface{}{
			"params": map[string]interface{}{"name": toList[0].Name},
			"email":  toList[0].Address,
		})
	}

	if d := cmp.Diff(recipients, actualMeta); d != "" {
		t.Fatalf("Unexpected delivery meta: %s", d)
	}
}
