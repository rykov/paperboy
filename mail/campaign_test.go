package mail

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
)

func TestCampaignEndToEnd(t *testing.T) {
	// Setup virtual filesystem
	memFs := afero.NewMemMapFs()

	// Create basic directory structure
	memFs.MkdirAll("content", 0755)
	memFs.MkdirAll("lists", 0755)
	memFs.MkdirAll("layouts", 0755)

	// Create a basic email template with frontmatter
	emailContent := `---
subject: "Test Newsletter"
from: "test@example.com"
---

# Hello {{ .Recipient.Name }}!

Welcome to our newsletter. This is a test campaign.

Best regards,
The Team`

	afero.WriteFile(memFs, "content/newsletter.md", []byte(emailContent), 0644)

	// Create a basic recipient list
	recipientList := `- name: "John Doe"
  email: "john@example.com"
- name: "Jane Smith"
  email: "jane@example.com"`

	afero.WriteFile(memFs, "lists/subscribers.yaml", []byte(recipientList), 0644)

	// Create default layouts
	htmlLayout := `<html>
<head><title>{{ .Subject }}</title></head>
<body>
{{ .Content }}
<hr>
<p><a href="{{ .UnsubscribeURL }}">Unsubscribe</a></p>
<p>{{ .Address }}</p>
</body>
</html>`

	textLayout := `{{ .Content }}

---
Unsubscribe: {{ .UnsubscribeURL }}
{{ .Address }}`

	afero.WriteFile(memFs, "layouts/_default.html", []byte(htmlLayout), 0644)
	afero.WriteFile(memFs, "layouts/_default.text", []byte(textLayout), 0644)

	// Load configuration using the new config system
	cfg, err := config.LoadConfigFs(t.Context(), memFs)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Configure for dry-run testing
	cfg.DryRun = true
	cfg.From = "newsletter@example.com"
	cfg.Address = "123 Main St, Anytown, USA"
	cfg.UnsubscribeURL = "https://example.com/unsubscribe?email={recipient.email}"
	cfg.ContentDir = "content"
	cfg.ListDir = "lists"
	cfg.LayoutDir = "layouts"
	cfg.Workers = 1
	cfg.SendRate = 0

	// Load campaign using new API
	campaign, err := LoadCampaign(cfg, "newsletter", "subscribers")
	if err != nil {
		t.Fatalf("Failed to load campaign: %v", err)
	}

	// Verify campaign loaded correctly
	if campaign.ID != "newsletter" {
		t.Errorf("Expected campaign ID 'newsletter', got '%s'", campaign.ID)
	}

	if len(campaign.Recipients) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(campaign.Recipients))
	}

	if campaign.EmailMeta.Subject() != "Test Newsletter" {
		t.Errorf("Expected subject 'Test Newsletter', got '%s'", campaign.EmailMeta.Subject())
	}

	if campaign.EmailMeta.From != "test@example.com" {
		t.Errorf("Expected from 'test@example.com', got '%s'", campaign.EmailMeta.From)
	}

	// Test message generation for first recipient
	message, err := campaign.MessageFor(0)
	if err != nil {
		t.Fatalf("Failed to generate message: %v", err)
	}

	// Verify message content by checking the raw message
	var buf bytes.Buffer
	if _, err := message.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	msgContent := buf.String()
	t.Logf("Message content: %s", msgContent) // Debug output
	if !strings.Contains(msgContent, "john@example.com") {
		t.Errorf("Expected email john@example.com in message content")
	}

	if !strings.Contains(msgContent, "Test Newsletter") {
		t.Errorf("Expected Subject 'Test Newsletter' in message content")
	}

	if !strings.Contains(msgContent, "test@example.com") {
		t.Errorf("Expected From email 'test@example.com' in message content")
	}

	// Test message generation for second recipient
	message2, err := campaign.MessageFor(1)
	if err != nil {
		t.Fatalf("Failed to generate message for second recipient: %v", err)
	}

	var buf2 bytes.Buffer
	if _, err := message2.WriteTo(&buf2); err != nil {
		t.Fatalf("Failed to write message2: %v", err)
	}

	msgContent2 := buf2.String()
	if !strings.Contains(msgContent2, "jane@example.com") {
		t.Errorf("Expected email jane@example.com in message2 content")
	}

	// Test full campaign send (dry-run) using new API
	err = SendCampaign(cfg, campaign)
	if err != nil {
		t.Fatalf("Failed to send campaign: %v", err)
	}

	t.Log("Campaign test completed successfully with dry-run sender")
}
