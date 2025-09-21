package mail

import (
	"github.com/google/go-cmp/cmp"
	"github.com/rykov/paperboy/config"
	"testing"
)

func TestContextFlattenMap(t *testing.T) {
	input := map[string]interface{}{
		"Level1a": "1",
		"Level1b": map[string]interface{}{
			"Level2a": "2",
			"Level2b": map[string]interface{}{
				"Level3": "3",
			},
		},
	}

	expected := map[string]interface{}{
		"Level1a":                "1",
		"Level1b.Level2a":        "2",
		"Level1b.Level2b.Level3": "3",
	}

	if out := flattenMap(input); !cmp.Equal(out, expected) {
		t.Errorf("Output mismatch:\nExpected:%v\nActual:%v", expected, out)
	}
}

func TestNewCampaignWithAttachments(t *testing.T) {
	cfg := &config.AConfig{}

	tests := []struct {
		name        string
		data        map[string]interface{}
		expected    []string
		description string
	}{
		{
			name: "single_attachment_string",
			data: map[string]interface{}{
				"subject":     "Test Subject",
				"from":        "test@example.com",
				"attachments": "document.pdf",
			},
			expected:    []string{"document.pdf"},
			description: "single attachment as string",
		},
		{
			name: "multiple_attachments_slice",
			data: map[string]interface{}{
				"subject":     "Test Subject",
				"from":        "test@example.com",
				"attachments": []string{"document.pdf", "image.png", "data.csv"},
			},
			expected:    []string{"document.pdf", "image.png", "data.csv"},
			description: "multiple attachments as string slice",
		},
		{
			name: "multiple_attachments_interface_slice",
			data: map[string]interface{}{
				"subject": "Test Subject",
				"from":    "test@example.com",
				"attachments": []interface{}{
					"document.pdf",
					"image.png",
					"data.csv",
				},
			},
			expected:    []string{"document.pdf", "image.png", "data.csv"},
			description: "multiple attachments as interface slice",
		},
		{
			name: "no_attachments",
			data: map[string]interface{}{
				"subject": "Test Subject",
				"from":    "test@example.com",
			},
			expected:    nil,
			description: "no attachments field",
		},
		{
			name: "empty_attachments_string",
			data: map[string]interface{}{
				"subject":     "Test Subject",
				"from":        "test@example.com",
				"attachments": "",
			},
			expected:    []string{},
			description: "empty string attachment",
		},
		{
			name: "empty_attachments_slice",
			data: map[string]interface{}{
				"subject":     "Test Subject",
				"from":        "test@example.com",
				"attachments": []string{},
			},
			expected:    []string{},
			description: "empty attachment slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign := newCampaign(cfg, tt.data)

			if !cmp.Equal(campaign.attachments, tt.expected) {
				t.Errorf("Attachments mismatch for %s:\nExpected: %v\nActual: %v",
					tt.description, tt.expected, campaign.attachments)
			}

			// Verify attachments field is removed from Params
			if _, exists := campaign.Params["attachments"]; exists {
				t.Errorf("Expected 'attachments' to be removed from Params")
			}

			// Verify other fields are processed correctly
			if campaign.subject != "Test Subject" {
				t.Errorf("Expected subject 'Test Subject', got '%s'", campaign.subject)
			}
			if campaign.From != "test@example.com" {
				t.Errorf("Expected from 'test@example.com', got '%s'", campaign.From)
			}
		})
	}
}

func TestNewCampaignAttachmentEdgeCases(t *testing.T) {
	cfg := &config.AConfig{}

	tests := []struct {
		name        string
		data        map[string]interface{}
		expected    []string
		description string
	}{
		{
			name: "invalid_attachment_type",
			data: map[string]interface{}{
				"subject":     "Test Subject",
				"attachments": 12345, // Invalid type
			},
			expected:    []string{"12345"}, // cast.ToString converts numbers to strings
			description: "invalid attachment type (integer)",
		},
		{
			name: "mixed_type_slice",
			data: map[string]interface{}{
				"subject": "Test Subject",
				"attachments": []interface{}{
					"valid.pdf",
					123, // Invalid type in slice
					"valid.png",
				},
			},
			expected:    []string{"valid.pdf", "123", "valid.png"}, // cast.ToStringSliceE converts numbers to strings
			description: "mixed types in attachment slice",
		},
		{
			name: "nil_attachments",
			data: map[string]interface{}{
				"subject":     "Test Subject",
				"attachments": nil,
			},
			expected:    nil,
			description: "nil attachments value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign := newCampaign(cfg, tt.data)

			if !cmp.Equal(campaign.attachments, tt.expected) {
				t.Errorf("Attachments mismatch for %s:\nExpected: %v\nActual: %v",
					tt.description, tt.expected, campaign.attachments)
			}
		})
	}
}
