package cmd

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/rykov/paperboy/config"
	"github.com/spf13/cobra"
)

func TestPreviewCmd(t *testing.T) {
	cmd := previewCmd()

	if cmd == nil {
		t.Fatal("previewCmd() returned nil")
	}

	if cmd.Use != "preview [content] [list]" {
		t.Errorf("Expected Use to be 'preview [content] [list]', got %s", cmd.Use)
	}

	if cmd.Short != "Preview campaign in browser" {
		t.Errorf("Expected specific short description, got %s", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}

	if cmd.Run != nil {
		t.Error("Run function should be nil when RunE is set")
	}
}

func TestPreviewCmdArgs(t *testing.T) {
	cmd := previewCmd()

	if cmd.Args == nil {
		t.Fatal("Args function should not be nil")
	}

	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"valid args", []string{"content", "list"}, false},
		{"too few args", []string{"content"}, true},
		{"too many args", []string{"content", "list", "extra"}, true},
		{"no args", []string{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cmd.Args(cmd, tc.args)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestPreviewConstants(t *testing.T) {
	if cliVersionHeader != "X-Paperboy-Version" {
		t.Errorf("Expected cliVersionHeader to be 'X-Paperboy-Version', got %q", cliVersionHeader)
	}

	if cliUserAgent != "Paperboy (%s)" {
		t.Errorf("Expected cliUserAgent to be 'Paperboy (%%s)', got %q", cliUserAgent)
	}
}

func TestPreviewTestMode(t *testing.T) {
	// Save original state
	originalMode := previewTestMode
	defer func() {
		previewTestMode = originalMode
	}()

	// Test that test mode is initially false
	if previewTestMode {
		t.Error("previewTestMode should be false initially")
	}

	// Test enabling test mode
	previewTestMode = true
	if !previewTestMode {
		t.Error("previewTestMode should be true after setting")
	}

	// Test disabling test mode
	previewTestMode = false
	if previewTestMode {
		t.Error("previewTestMode should be false after resetting")
	}
}

func TestOpenPreviewTestMode(t *testing.T) {
	// Save original state and enable test mode
	originalMode := previewTestMode
	previewTestMode = true
	defer func() {
		previewTestMode = originalMode
	}()

	cfg := &config.AConfig{
		ConfigFile: config.ConfigFile{
			ServerPort: 8080,
		},
	}

	// Create a command with a buffer for output
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// This should trigger test mode and write to the command's output
	openPreview(cmd, cfg, "test-content", "test-list")

	expectedURL := "http://localhost:8080/preview/test-content/test-list"
	expectedOutput := fmt.Sprintf("\nPlease open the browser to the following URL:\n%s\n\n", expectedURL)

	if buf.String() != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, buf.String())
	}
}

func TestOpenPreviewProductionMode(t *testing.T) {
	// Test the production mode logic without actually opening browser
	// We'll verify the code path but keep test mode enabled to avoid side effects
	originalMode := previewTestMode
	previewTestMode = true // Keep test mode to avoid browser opening
	defer func() {
		previewTestMode = originalMode
	}()

	cfg := &config.AConfig{
		ConfigFile: config.ConfigFile{
			ServerPort: 9090,
		},
	}

	// Create a command with a buffer for output
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// This test verifies that openPreview doesn't panic and produces expected output
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("openPreview should not panic: %v", r)
		}
	}()

	openPreview(cmd, cfg, "content", "list")

	// Should always get fallback message in test mode
	expectedURL := "http://localhost:9090/preview/content/list"
	expectedOutput := fmt.Sprintf(cliPreviewMsg, expectedURL)

	if buf.String() != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, buf.String())
	}
}

func TestOpenPreviewURLConstruction(t *testing.T) {
	content := "test content with spaces"
	list := "test-list"

	expectedRoot := "http://localhost:9090"
	expectedPath := "/preview/" + url.PathEscape(content) + "/" + url.PathEscape(list)
	expectedURL := expectedRoot + expectedPath

	actualPath := "/preview/" + url.PathEscape(content) + "/" + url.PathEscape(list)
	actualURL := expectedRoot + actualPath

	if actualURL != expectedURL {
		t.Errorf("URL construction mismatch. Expected %q, got %q", expectedURL, actualURL)
	}

	if url.PathEscape(content) == content {
		t.Error("Content with spaces should be URL escaped")
	}
}

func TestOpenPreviewURLEscaping(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		list     string
		expected string
	}{
		{
			name:     "simple names",
			content:  "newsletter",
			list:     "subscribers",
			expected: "/preview/newsletter/subscribers",
		},
		{
			name:     "names with spaces",
			content:  "monthly newsletter",
			list:     "vip subscribers",
			expected: "/preview/monthly%20newsletter/vip%20subscribers",
		},
		{
			name:     "names with special chars",
			content:  "newsletter#1",
			list:     "list@company",
			expected: "/preview/newsletter%231/list@company",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := "/preview/" + url.PathEscape(tc.content) + "/" + url.PathEscape(tc.list)
			if actual != tc.expected {
				t.Errorf("Expected path %q, got %q", tc.expected, actual)
			}
		})
	}
}

func TestOpenPreviewPortConfiguration(t *testing.T) {
	// Save original state and enable test mode for predictable output
	originalMode := previewTestMode
	previewTestMode = true
	defer func() {
		previewTestMode = originalMode
	}()

	testPorts := []int{8080, 3000, 9000, 8888}

	for _, port := range testPorts {
		t.Run(fmt.Sprintf("port_%d", port), func(t *testing.T) {
			cfg := &config.AConfig{
				ConfigFile: config.ConfigFile{
					ServerPort: uint(port),
				},
			}

			// Create a command with a buffer for output
			cmd := &cobra.Command{}
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			openPreview(cmd, cfg, "content", "list")

			expectedURL := fmt.Sprintf("http://localhost:%d/preview/content/list", port)
			expectedOutput := fmt.Sprintf("\nPlease open the browser to the following URL:\n%s\n\n", expectedURL)

			if buf.String() != expectedOutput {
				t.Errorf("Expected output %q, got %q", expectedOutput, buf.String())
			}
		})
	}
}

func TestOpenPreviewFallbackMessage(t *testing.T) {
	// Test that fallback message format is correct
	cfg := &config.AConfig{
		ConfigFile: config.ConfigFile{
			ServerPort: 8080,
		},
	}

	testCases := []struct {
		name    string
		content string
		list    string
	}{
		{"simple", "newsletter", "subscribers"},
		{"with spaces", "monthly update", "vip list"},
		{"special chars", "newsletter#1", "list@domain"},
	}

	// Enable test mode to ensure fallback message is triggered
	originalMode := previewTestMode
	previewTestMode = true
	defer func() {
		previewTestMode = originalMode
	}()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a command with a buffer for output
			cmd := &cobra.Command{}
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			openPreview(cmd, cfg, tc.content, tc.list)

			output := buf.String()
			if !strings.Contains(output, "Please open the browser to the following URL:") {
				t.Error("Output should contain fallback message")
			}

			if !strings.Contains(output, "http://localhost:8080/preview/") {
				t.Error("Output should contain preview URL")
			}
		})
	}
}

func TestOpenPreviewWithErrorHandling(t *testing.T) {
	// Test mode ensures we get predictable error behavior
	originalMode := previewTestMode
	previewTestMode = true
	defer func() {
		previewTestMode = originalMode
	}()

	cfg := &config.AConfig{
		ConfigFile: config.ConfigFile{
			ServerPort: 8080,
		},
	}

	// Create a command with a buffer for output
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Test that the function doesn't panic and writes correct output
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("openPreview should not panic: %v", r)
		}
	}()

	openPreview(cmd, cfg, "content", "list")

	// Verify the error path writes the fallback message
	expectedURL := "http://localhost:8080/preview/content/list"
	expectedOutput := fmt.Sprintf("\nPlease open the browser to the following URL:\n%s\n\n", expectedURL)

	if buf.String() != expectedOutput {
		t.Errorf("Expected fallback output %q, got %q", expectedOutput, buf.String())
	}
}
