package mail

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rykov/paperboy/config"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Default schema applied when no custom schema exists and no custom "to" template is defined
// Ensures recipients have at least an email address while allowing additional properties
const defaultRecipientSchema = `{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"title": "DefaultRecipient",
	"type": "object",
	"required": ["email"],
	"additionalProperties": true,
	"properties": {
		"name": {"type": "string"},
		"email": {"type": "string", "format": "email"}
	}
}`

// compileDefaultSchema compiles the default recipient schema
func compileDefaultSchema() (*jsonschema.Schema, error) {
	schemaName := "default.schema"
	r := strings.NewReader(defaultRecipientSchema)
	return compileSchema(r, schemaName, schemaName)
}

// loadRecipientSchema loads a JSON schema file for recipient parameter validation
// For a campaign "newsletter.md", it looks for "newsletter.schema"
func loadRecipientSchema(appFs *config.Fs, tmplID string) (*jsonschema.Schema, error) {
	// Build schema file path using FindSchemaPath or return no schema
	schemaPath := appFs.FindSchemaPath(tmplID)
	if schemaPath == "" {
		return nil, nil
	}

	// Open schema file reader
	schemaFile, err := appFs.Open(schemaPath)
	if err != nil {
		return nil, wrapSchemaError(err, "read", schemaPath)
	}
	defer schemaFile.Close()

	schemaName := "schema://" + schemaPath
	return compileSchema(schemaFile, schemaName, schemaPath)
}

func compileSchema(r io.Reader, schemaName, schemaPath string) (*jsonschema.Schema, error) {
	// Parse JSON schema content
	var schemaDoc any
	if err := json.NewDecoder(r).Decode(&schemaDoc); err != nil {
		return nil, wrapSchemaError(err, "parse JSON", schemaPath)
	}

	// Compile JSON schema
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat() // Force-enable format assertions
	if err := compiler.AddResource(schemaName, schemaDoc); err != nil {
		return nil, wrapSchemaError(err, "add schema resource", schemaPath)
	}

	schema, err := compiler.Compile(schemaName)
	if err != nil {
		return nil, wrapSchemaError(err, "compile", schemaPath)
	}

	return schema, nil
}

// wrapSchemaError wraps schema-related errors with consistent formatting
func wrapSchemaError(err error, operation, schemaPath string) error {
	return fmt.Errorf("failed to %s schema file %s: %w", operation, schemaPath, err)
}

// loadRecipientSchemaWithDefault loads a schema file if it exists, otherwise returns default schema
// if campaign has no custom "to" template defined
func loadRecipientSchemaWithDefault(appFs *config.Fs, tmplID string, campaign *ctxCampaign) (*jsonschema.Schema, error) {
	// First try to load custom schema
	schema, err := loadRecipientSchema(appFs, tmplID)
	if err != nil {
		return nil, err
	}

	// If custom schema exists, use it
	if schema != nil {
		return schema, nil
	}

	// If no custom schema and campaign has custom "to" template, skip validation
	if campaign != nil && strings.TrimSpace(campaign.to) != "" {
		return nil, nil
	}

	// Apply default schema
	return compileDefaultSchema()
}
