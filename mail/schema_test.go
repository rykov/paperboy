package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rykov/paperboy/config"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/afero"
)

// Test helper functions

// createTestEnvironment creates a test filesystem with config for schema testing
// If withAppFs is true, creates a full config with AppFs for campaign testing
func createTestEnvironment(t *testing.T, withAppFs bool) (*config.AConfig, *config.Fs, afero.Fs) {
	t.Helper()
	fs := afero.NewMemMapFs()
	cfg := &config.AConfig{
		ConfigFile: config.ConfigFile{
			ContentDir: "content",
			ListDir:    "lists",
		},
		Context: context.Background(),
	}
	appFs := &config.Fs{Config: cfg, Fs: fs}

	if withAppFs {
		cfg.AppFs = appFs
		return cfg, appFs, fs
	}
	return nil, appFs, fs
}

// writeTestFile writes content to a file path, failing the test on error
func writeTestFile(t *testing.T, fs afero.Fs, path, content string) {
	t.Helper()
	if err := afero.WriteFile(fs, path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
}

// assertValidationSuccess checks that no error occurred
func assertValidationSuccess(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// assertValidationFailure checks that an error occurred with expected message
func assertValidationFailure(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Error("Expected error but got none")
		return
	}
	if expectedMsg != "" && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain %q, got: %v", expectedMsg, err)
	}
}

// assertSchemaLoaded checks if schema was loaded as expected
func assertSchemaLoaded(t *testing.T, schema *jsonschema.Schema, shouldExist bool) {
	t.Helper()
	if shouldExist {
		if schema == nil {
			t.Error("Expected schema but got nil")
		}
	} else {
		if schema != nil {
			t.Error("Expected no schema but got one")
		}
	}
}

// Standard test content
const (
	validEmail = "john@example.com"
	validName  = "John Doe"

	testCampaignContent = `---
subject: "Newsletter"
from: "test@example.com"
---
# Newsletter Content`
)

// buildRecipient creates recipient YAML data with optional extra fields
func buildRecipient(email, name string, extra map[string]any) string {
	result := fmt.Sprintf(`- name: "%s"`, name)
	if email != "" {
		result += fmt.Sprintf(`
  email: "%s"`, email)
	}
	for k, v := range extra {
		result += fmt.Sprintf(`
  %s: "%v"`, k, v)
	}
	return result
}

// buildSchema creates JSON schema with required fields and properties
func buildSchema(required []string, properties map[string]any) string {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": true,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	if len(properties) > 0 {
		schema["properties"] = properties
	}
	data, _ := json.MarshalIndent(schema, "", "\t")
	return string(data)
}

func TestLoadRecipientSchema(t *testing.T) {
	tests := []struct {
		name          string
		tmplID        string
		schemaExists  bool
		schemaContent string
		expectError   bool
		expectSchema  bool
	}{
		{
			name:         "no schema file",
			tmplID:       "newsletter",
			schemaExists: false,
			expectError:  false,
			expectSchema: false,
		},
		{
			name:          "valid schema file",
			tmplID:        "newsletter",
			schemaExists:  true,
			schemaContent: buildSchema([]string{"company"}, map[string]any{"company": map[string]any{"type": "string"}}),
			expectError:   false,
			expectSchema:  true,
		},
		{
			name:          "invalid schema file",
			tmplID:        "newsletter",
			schemaExists:  true,
			schemaContent: `invalid json`,
			expectError:   true,
			expectSchema:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appFs, fs := createTestEnvironment(t, false)

			// Create schema file if needed
			if tt.schemaExists {
				writeTestFile(t, fs, "content/newsletter.schema", tt.schemaContent)
			}

			schema, err := loadRecipientSchema(appFs, tt.tmplID)

			if tt.expectError {
				assertValidationFailure(t, err, "")
			} else {
				assertValidationSuccess(t, err)
			}

			assertSchemaLoaded(t, schema, tt.expectSchema)
		})
	}
}

func TestRecipientSchemaValidation(t *testing.T) {
	tests := []struct {
		name          string
		schemaType    string // "none", "custom", "default", "with-custom-to", "format"
		recipientData string
		schemaContent string
		customTo      string
		expectError   bool
		errorMsg      string
	}{
		// Custom schema tests
		{
			name:          "valid recipients with custom schema",
			schemaType:    "custom",
			recipientData: buildRecipient(validEmail, validName, map[string]any{"company": "Acme Inc"}),
			schemaContent: buildSchema([]string{"company"}, map[string]any{"company": map[string]any{"type": "string"}}),
			expectError:   false,
		},
		{
			name:          "invalid recipients missing required field",
			schemaType:    "custom",
			recipientData: buildRecipient(validEmail, validName, map[string]any{"role": "Developer"}),
			schemaContent: buildSchema([]string{"company"}, map[string]any{"company": map[string]any{"type": "string"}}),
			expectError:   true,
			errorMsg:      "recipient 0: jsonschema validation failed",
		},

		// Default schema tests
		{
			name:          "valid recipient with email (default schema)",
			schemaType:    "default",
			recipientData: buildRecipient(validEmail, validName, nil),
			expectError:   false,
		},
		{
			name:          "invalid recipient without email (default schema)",
			schemaType:    "default",
			recipientData: buildRecipient("", validName, nil),
			expectError:   true,
			errorMsg:      "missing property 'email'",
		},
		{
			name:          "valid recipient with additional properties (default schema)",
			schemaType:    "default",
			recipientData: buildRecipient(validEmail, validName, map[string]any{"company": "Acme Inc", "role": "Developer"}),
			expectError:   false,
		},

		// Email format validation tests
		{
			name:          "valid email format",
			schemaType:    "format",
			recipientData: buildRecipient(validEmail, validName, nil),
			expectError:   false,
		},
		{
			name:          "invalid email format - no @ symbol",
			schemaType:    "format",
			recipientData: buildRecipient("notanemail", validName, nil),
			expectError:   true,
			errorMsg:      "missing @",
		},
		{
			name:          "invalid email format - missing domain",
			schemaType:    "format",
			recipientData: buildRecipient("user@", validName, nil),
			expectError:   true,
			errorMsg:      "invalid domain",
		},

		// Custom "to" template tests (should skip validation)
		{
			name:          "no validation with custom to template",
			schemaType:    "with-custom-to",
			recipientData: buildRecipient("", validName, nil), // Missing email, should not cause error
			customTo:      "{{ .Recipient.Name }} <custom@example.com>",
			expectError:   false,
		},
		{
			name:          "no validation with custom to template - empty params",
			schemaType:    "with-custom-to",
			recipientData: buildRecipient("", "Test", nil),
			customTo:      "test@example.com",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appFs, fs := createTestEnvironment(t, false)

			// Create recipient file
			writeTestFile(t, fs, "lists/test.yaml", tt.recipientData)

			// Create schema file if provided
			if tt.schemaContent != "" {
				writeTestFile(t, fs, "content/test.schema", tt.schemaContent)
			}

			// Parse recipients
			recipients, err := parseRecipients(appFs, "lists/test.yaml")
			if err != nil {
				t.Fatalf("Failed to parse recipients: %v", err)
			}

			// Create campaign with or without custom "to" template
			campaign := &ctxCampaign{to: tt.customTo}

			// Test schema verification
			err = verifyRecipientSchema(appFs, "test.md", campaign, recipients)

			if tt.expectError {
				assertValidationFailure(t, err, tt.errorMsg)
			} else {
				assertValidationSuccess(t, err)
			}
		})
	}
}

func TestVerifyCampaignWithRecipientSchemaError(t *testing.T) {
	cfg, _, fs := createTestEnvironment(t, true)

	// Create test files
	writeTestFile(t, fs, "content/newsletter.md", testCampaignContent)
	writeTestFile(t, fs, "lists/subscribers.yaml", buildRecipient(validEmail, validName, map[string]any{"role": "Developer"}))
	writeTestFile(t, fs, "content/newsletter.schema", buildSchema([]string{"company"}, map[string]any{"company": map[string]any{"type": "string"}}))

	// Test VerifyCampaign which should catch the schema validation error
	err := VerifyCampaign(cfg, "newsletter", "subscribers")

	assertValidationFailure(t, err, "recipient schema validation failed")
}

func TestLoadCampaignDoesNotValidateSchema(t *testing.T) {
	// Test that normal LoadCampaign (used by send) does NOT validate schemas
	cfg, _, fs := createTestEnvironment(t, true)

	// Create files - recipients with missing required field according to schema
	writeTestFile(t, fs, "content/newsletter.md", testCampaignContent)
	writeTestFile(t, fs, "lists/subscribers.yaml", buildRecipient(validEmail, validName, map[string]any{"role": "Developer"}))
	writeTestFile(t, fs, "content/newsletter.schema", buildSchema([]string{"company"}, map[string]any{"company": map[string]any{"type": "string"}}))

	// LoadCampaign should succeed even with invalid recipients (schema not checked)
	campaign, err := LoadCampaign(cfg, "newsletter", "subscribers")

	assertValidationSuccess(t, err)

	if campaign == nil {
		t.Error("Expected campaign to load successfully")
		return
	}

	if len(campaign.Recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(campaign.Recipients))
	}

	// Verify the recipient data was loaded correctly
	if e := campaign.Recipients[0].Email(); e != validEmail {
		t.Errorf("Expected recipient email %s, got %s", validEmail, e)
	}

	// Verify the invalid param (missing 'company') is still present
	if role, exists := (*campaign.Recipients[0])["role"]; !exists || role != "Developer" {
		t.Error("Expected recipient role parameter to be loaded")
	}
}
