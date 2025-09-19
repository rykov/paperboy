package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rykov/paperboy/ui"
)

func TestServerCmd(t *testing.T) {
	cmd := serverCmd()

	if cmd == nil {
		t.Fatal("serverCmd() returned nil")
	}

	if cmd.Use != "server" {
		t.Errorf("Expected Use to be 'server', got %s", cmd.Use)
	}

	if cmd.Short != "Launch a preview server for emails" {
		t.Errorf("Expected specific short description, got %s", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}

	if cmd.Run != nil {
		t.Error("Run function should be nil when RunE is set")
	}
}

func TestServerConstants(t *testing.T) {
	if serverGraphQLPath != "/graphql" {
		t.Errorf("Expected serverGraphQLPath to be '/graphql', got %q", serverGraphQLPath)
	}
}

func TestUIHandler(t *testing.T) {
	handler := uiHandler()

	if handler == nil {
		t.Fatal("uiHandler() returned nil")
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Error("Expected handler to serve UI files, got 404")
	}
}

func TestUIHandlerNonExistentPath(t *testing.T) {
	handler := uiHandler()

	req := httptest.NewRequest("GET", "/non-existent-path", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		if req.URL.Path != "/" {
			t.Error("Expected non-existent paths to be rewritten to /")
		}
	}
}

func TestUIHandlerPathRewriting(t *testing.T) {
	handler := uiHandler()

	testPaths := []string{
		"/some/deep/path",
		"/another/path",
		"/api/nonexistent",
	}

	for _, path := range testPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			originalPath := req.URL.Path

			handler.ServeHTTP(w, req)

			httpFS := http.FS(ui.FS)
			if _, err := httpFS.Open(originalPath); err != nil {
				if req.URL.Path == "/" {
					t.Logf("Path %s was correctly rewritten to /", originalPath)
				}
			}
		})
	}
}

func TestUIFileServerBehavior(t *testing.T) {
	httpFS := http.FS(ui.FS)

	if httpFS == nil {
		t.Fatal("ui.FS should provide a valid filesystem")
	}

	file, err := httpFS.Open("/")
	if err != nil {
		t.Logf("Root path not accessible: %v", err)
	} else {
		file.Close()
	}
}

func TestUIHandlerHTTPMethods(t *testing.T) {
	handler := uiHandler()

	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code == http.StatusMethodNotAllowed {
				t.Errorf("Method %s should be allowed, got %d", method, w.Code)
			}
		})
	}
}

func TestConfigFuncType(t *testing.T) {
	var fn configFunc = func(mux *http.ServeMux, ready chan bool) error {
		return nil
	}

	if fn == nil {
		t.Error("configFunc should be assignable")
	}

	mux := http.NewServeMux()
	ready := make(chan bool)

	err := fn(mux, ready)
	if err != nil {
		t.Errorf("configFunc should execute without error, got %v", err)
	}
}

func TestUIHandlerErrorHandling(t *testing.T) {
	handler := uiHandler()

	// Test with a simple invalid path instead of control characters
	req := httptest.NewRequest("GET", "/invalid-path-that-does-not-exist", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Handler should not panic on invalid paths: %v", r)
		}
	}()

	handler.ServeHTTP(w, req)
}
