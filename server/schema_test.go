package server

import (
	"github.com/jordan-wright/email"
	"github.com/neelance/graphql-go"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"

	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	mail.SetFs(afero.NewMemMapFs())
	os.Exit(m.Run())
}

func TestSchemaBasicQuery(t *testing.T) {
	var fs = mail.AppFs

	afero.WriteFile(fs, "config.toml", []byte(""), 0644)
	afero.WriteFile(fs, fs.ContentPath("c1.md"), []byte("# Hello"), 0644)
	afero.WriteFile(fs, fs.ListPath("r1.yaml"), []byte(`---
- email: ex@example.org
`), 0644)

	schema := graphql.MustParseSchema(schemaText, &Resolver{})
	response := schema.Exec(context.TODO(), `{
		renderOne(content: "c1", recipient: "r1#0") {
			rawMessage
			text
			html
		}
	}`, "", map[string]interface{}{})

	if errs := response.Errors; len(errs) > 0 {
		t.Fatalf("GraphQL errors %+v", errs)
	}

	resp := struct {
		RenderOne struct {
			RawMessage string
			Text       string
			HTML       string
		}
	}{}

	if err := json.Unmarshal(response.Data, &resp); err != nil {
		t.Fatalf("JSON unmarshal error: %s", err)
	}

	// Text extracted text and html parts
	if s := resp.RenderOne.Text; s != "# Hello" {
		t.Errorf("Invalid text: %s", s)
	}
	if s := resp.RenderOne.HTML; !strings.Contains(s, "<h1>Hello</h1>") {
		t.Errorf("Invalid html: %s", s)
	}

	// Parse raw message to verify all the fields
	em, err := email.NewEmailFromReader(strings.NewReader(resp.RenderOne.RawMessage))
	if err != nil {
		t.Fatalf("Invalid RawMessage: %s", err)
	}
	if r := em.To; !reflect.DeepEqual(r, []string{"ex@example.org"}) {
		t.Errorf("Invalid email.To: %+v", r)
	}
	if s := string(em.HTML); s != resp.RenderOne.HTML {
		t.Errorf("Invalid RawMessage HTML: %s", s)
	}
	if s := string(em.Text); s != resp.RenderOne.Text {
		t.Errorf("Invalid RawMessage Text: %s", s)
	}
}
