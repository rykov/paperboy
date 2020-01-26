package mail

import (
	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"

	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	config.SetFs(afero.NewMemMapFs())
	os.Exit(m.Run())
}

func NewTestConfig(t *testing.T) *config.AConfig {
	cfg := config.NewConfig(afero.NewMemMapFs())
	expect := reflect.TypeOf((*afero.MemMapFs)(nil))
	if expect != reflect.TypeOf(cfg.AppFs.Fs) {
		t.Errorf("AppFs should be MemMapFs - check setup")
	}
	return cfg
}

func TestInlineStylesheetsSuccess(t *testing.T) {
	layoutPath := "/inline-test/file.html"

	// Campaign configured for testing
	c := &Campaign{Config: NewTestConfig(t)}
	appFs := c.Config.AppFs

	// Regular no stylesheet template
	expect := "<body>Hello World</body>"
	out, err := c.inlineStylesheets(layoutPath, expect)
	if err != nil || !strings.Contains(out, expect) {
		t.Errorf("Basic no-inline failed (%s): %s", err, out)
	}

	// Regular inline stylesheet
	out, err = c.inlineStylesheets(layoutPath, `
		<style> h1 { color: #123; }</style>
		<h1>Hello World</h1>
	`)
	expect = "<h1 style=\"color: #123;\">Hello World"
	if err != nil || !strings.Contains(out, expect) {
		t.Errorf("Inlining <style> failed: %q doesn't contain %q", out, expect)
	}

	// Regular from file stylesheet
	testCSS := filepath.Join(filepath.Dir(layoutPath), "test.css")
	afero.WriteFile(appFs, testCSS, []byte("h1 { color: #321; }"), 0644)
	out, err = c.inlineStylesheets(layoutPath, `
		<link rel="stylesheet" href="test.css"/>
		<h1>Hello World</h1>
	`)
	expect = "<h1 style=\"color: #321;\">Hello World"
	if err != nil || !strings.Contains(out, expect) {
		t.Errorf("Inlining <style> failed: %q doesn't contain %q", out, expect)
	}

	// Ignore <link> tags that's not a stylesheet
	expect = "<link rel=\"alternate\" href=\"test.css\"/>"
	out, err = c.inlineStylesheets(layoutPath, expect+`<h1>Hello World</h1>`)
	if err != nil || !strings.Contains(out, expect) {
		t.Errorf("Should not inline non-stylesheet <link>: %q contains %q", out, expect)
	}
}

func TestInlineStylesheetsFailure(t *testing.T) {
	c := &Campaign{Config: NewTestConfig(t)}
	layoutPath := "/inline-test/file.html"

	// Should fail if no file specified
	_, err := c.inlineStylesheets(layoutPath, `
		<link rel="stylesheet"/>
		<h1>Hello World</h1>
	`)
	if err == nil || !strings.Contains(err.Error(), "No href") {
		t.Errorf("Should output an error if no href: %s", err)
	}

	// Should fail if file doesn't exist
	_, err = c.inlineStylesheets(layoutPath, `
		<link rel="stylesheet" href="not-here.css"/>
		<h1>Hello World</h1>
	`)
	if err == nil || !strings.Contains(err.Error(), "file does not exist") {
		t.Errorf("Should output an error if no file: %s", err)
	}
}
