package client

import (
	"github.com/rykov/paperboy/server"

	"archive/zip"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSendIntegration(t *testing.T) {
	// 1) create temp dir with files
	dir := t.TempDir()
	tempFiles := maps.Clone(expected)
	tempFiles["baz.skip"] = "skip this plz"
	for name, content := range tempFiles {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("writing %q: %v", path, err)
		}
	}

	// 2) spin up the server
	h := server.MustSchemaHandler(schemaSDL, &testResolver{})
	srv := httptest.NewServer(server.WithMiddleware(h, nil))
	defer srv.Close()

	// 3) test various client requests
	args := SendArgs{ProjectPath: dir, Campaign: "testCampaign", List: "testList"}
	args.ProjectIgnores = []string{"*.skip"} // Test ignoring files
	cli := New(context.Background(), srv.URL)
	if err := cli.Send(args); err != nil {
		t.Fatalf("client.Send failed: %v", err)
	}

	args.List = "testError"
	if err := cli.Send(args); err == nil {
		t.Errorf("Expected server to error, got success")
	} else if a, e := err.Error(), "server error: testError"; a != e {
		t.Errorf("Expected error %q, got %q", e, a)
	}

	args.List = "testPanic"
	if err := cli.Send(args); err == nil {
		t.Errorf("Expected server to error, got success")
	} else if a, e := err.Error(), "server error: panic occurred: testPanic"; a != e {
		t.Errorf("Expected error %q, got %q", e, a)
	}
}

// minimal test‐only schema & resolver:
const schemaSDL = `
  schema { query: Query mutation: Mutation }
  type Query {}
  type Mutation {
    sendCampaign(campaign: String!, list: String!): Boolean!
  }
`

// expected files and their contents
var expected = map[string]string{
	"foo.txt": "hello foo",
	"bar.txt": "hello bar",
}

type testResolver struct{}

// resolver signature with context so we can pull the zip back out
func (r *testResolver) SendCampaign(ctx context.Context, args struct {
	Campaign string
	List     string
}) (bool, error) {
	if l := args.List; l == "testError" {
		return false, fmt.Errorf("%s", l)
	} else if l == "testPanic" {
		panic(l)
	}

	f, ok := server.RequestZipFile(ctx)
	if !ok || f == nil {
		return false, fmt.Errorf("zip file not found in context")
	}
	defer f.Close()

	// figure out its size
	info, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("stat temp file: %w", err)
	}

	// open it as a zip.Reader
	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		return false, fmt.Errorf("open zip: %w", err)
	}

	// track which files we saw
	seen := make(map[string]bool)

	for _, entry := range zr.File {
		rc, err := entry.Open()
		if err != nil {
			return false, fmt.Errorf("open entry %q: %w", entry.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return false, fmt.Errorf("read entry %q: %w", entry.Name, err)
		}

		want, exists := expected[entry.Name]
		if !exists {
			return false, fmt.Errorf("unexpected file %q in zip", entry.Name)
		}
		if string(data) != want {
			return false, fmt.Errorf("file %q contents = %q; want %q", entry.Name, data, want)
		}
		seen[entry.Name] = true
	}

	// make sure we saw them all
	for name := range expected {
		if !seen[name] {
			return false, fmt.Errorf("expected file %q but not found in zip", name)
		}
	}

	return true, nil
}
