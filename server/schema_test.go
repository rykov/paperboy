package server

import (
	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/jordan-wright/email"
	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRenderOneQuery(t *testing.T) {
	cfg, fs := newTestConfigAndFs(t)
	afero.WriteFile(fs, fs.ContentPath("c1.md"), []byte("# Hello"), 0644)
	afero.WriteFile(fs, fs.ListPath("r1.yaml"), []byte(`---
- email: ex@example.org
`), 0644)

	response := issueGraphQLQuery(cfg, `{
		renderOne(content: "c1", recipient: "r1#0") {
			rawMessage
			text
			html
		}
	}`)

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
	if s := resp.RenderOne.Text; s != "\r\n  # Hello\r\n\r\n" {
		t.Errorf("Invalid text: %q", s)
	}
	if s := resp.RenderOne.HTML; !strings.Contains(s, "<h1 id=\"hello\">Hello</h1>") {
		t.Errorf("Invalid html: %q", s)
	}

	// Parse raw message to verify all the fields
	em, err := email.NewEmailFromReader(strings.NewReader(resp.RenderOne.RawMessage))
	if err != nil {
		t.Fatalf("Invalid RawMessage: %s", err)
	}
	if r := em.To; !cmp.Equal(r, []string{"ex@example.org"}) {
		t.Errorf("Invalid email.To: %+v", r)
	}
	if s := string(em.HTML); s != resp.RenderOne.HTML {
		t.Errorf("Invalid RawMessage HTML: %s", s)
	}
	if s := string(em.Text); s != resp.RenderOne.Text {
		t.Errorf("Invalid RawMessage Text: %s", s)
	}
}

func TestPaperboyInfoQuery(t *testing.T) {
	cfg, _ := newTestConfigAndFs(t)

	expected := &cfg.Build
	expected.BuildDate = time.Now().String()
	expected.Version = "1.2.3"

	response := issueGraphQLQuery(cfg, `{
		paperboyInfo {
			version
			buildDate
		}
	}`)

	if errs := response.Errors; len(errs) > 0 {
		t.Fatalf("GraphQL errors %+v", errs)
	}

	resp := struct {
		PaperboyInfo struct {
			Version   string
			BuildDate string
		}
	}{}

	if err := json.Unmarshal(response.Data, &resp); err != nil {
		t.Fatalf("JSON unmarshal error: %s", err)
	}

	actual := resp.PaperboyInfo
	if actual.Version != expected.Version {
		t.Errorf("Invalid version: %s", actual.Version)
	}
	if actual.BuildDate != expected.BuildDate {
		t.Errorf("Invalid buildDate: %s", actual.BuildDate)
	}
}

func issueGraphQLQuery(cfg *config.AConfig, query string) *graphql.Response {
	return issueGraphQL(cfg, query, map[string]interface{}{})
}

func issueGraphQL(cfg *config.AConfig, query string, vars map[string]interface{}) *graphql.Response {
	schema := graphql.MustParseSchema(schemaText, &Resolver{cfg: cfg})
	return schema.Exec(context.TODO(), query, "", vars)
}

func newTestConfigAndFs(t *testing.T) (*config.AConfig, *config.Fs) {
	fs := afero.NewMemMapFs()

	// FIXME: Viper's config loading from non-global
	// instance is broken, need to file an issue
	viper.SetFs(fs)

	// Write and load fake configuration
	afero.WriteFile(fs, "/config.toml", []byte(""), 0644)
	cfg, err := config.LoadConfigFs(t.Context(), fs)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that sender is in dryRun mode for testing
	cfg.From = "sender@example.com"
	cfg.DryRun = true
	cfg.Workers = 1

	return cfg, cfg.AppFs
}
