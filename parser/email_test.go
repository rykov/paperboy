package parser

import (
	"strings"
	"testing"
)

func TestReadFromYAMLFrontmatter(t *testing.T) {
	content := `---
subject: "Test Subject"
from: "test@example.com"
date: "2023-01-01"
---

# Hello World

This is the email content in **Markdown**.`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse email: %v", err)
	}

	// Test frontmatter extraction
	frontmatter := email.FrontMatter()
	if len(frontmatter) == 0 {
		t.Error("Frontmatter should not be empty")
	}

	if !strings.Contains(string(frontmatter), `subject: "Test Subject"`) {
		t.Errorf("Frontmatter should contain subject, got: %s", string(frontmatter))
	}

	// Test content extraction
	content = string(email.Content())
	expectedContent := "# Hello World\n\nThis is the email content in **Markdown**."
	if strings.TrimSpace(content) != expectedContent {
		t.Errorf("Content mismatch.\nExpected: %q\nGot: %q", expectedContent, strings.TrimSpace(content))
	}

	// Test metadata parsing
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	metaMap, ok := metadata.(map[string]interface{})
	if !ok {
		t.Fatal("Metadata should be a map")
	}

	if subject, exists := metaMap["subject"]; !exists || subject != "Test Subject" {
		t.Errorf("Expected subject 'Test Subject', got %v", subject)
	}

	if from, exists := metaMap["from"]; !exists || from != "test@example.com" {
		t.Errorf("Expected from 'test@example.com', got %v", from)
	}

	// Test renderability
	if !email.IsRenderable() {
		t.Error("Email with YAML frontmatter should be renderable")
	}
}

func TestReadFromTOMLFrontmatter(t *testing.T) {
	content := `+++
subject = "TOML Test Subject"
from = "toml@example.com"
tags = ["newsletter", "test"]
+++

Content with TOML frontmatter.`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse TOML email: %v", err)
	}

	// Test frontmatter extraction
	frontmatter := email.FrontMatter()
	if !strings.Contains(string(frontmatter), `subject = "TOML Test Subject"`) {
		t.Errorf("TOML frontmatter should contain subject, got: %s", string(frontmatter))
	}

	// Test content extraction
	content = strings.TrimSpace(string(email.Content()))
	expectedContent := "Content with TOML frontmatter."
	if content != expectedContent {
		t.Errorf("Content mismatch.\nExpected: %q\nGot: %q", expectedContent, content)
	}

	// Test metadata parsing
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to parse TOML metadata: %v", err)
	}

	metaMap, ok := metadata.(map[string]interface{})
	if !ok {
		t.Fatal("Metadata should be a map")
	}

	if subject, exists := metaMap["subject"]; !exists || subject != "TOML Test Subject" {
		t.Errorf("Expected subject 'TOML Test Subject', got %v", subject)
	}
}

func TestReadFromJSONFrontmatter(t *testing.T) {
	content := `{
  "subject": "JSON Test Subject",
  "from": "json@example.com",
  "priority": 1
}

JSON frontmatter content here.`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse JSON email: %v", err)
	}

	// Test metadata parsing
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to parse JSON metadata: %v", err)
	}

	metaMap, ok := metadata.(map[string]interface{})
	if !ok {
		t.Fatal("Metadata should be a map")
	}

	if subject, exists := metaMap["subject"]; !exists || subject != "JSON Test Subject" {
		t.Errorf("Expected subject 'JSON Test Subject', got %v", subject)
	}

	if priority, exists := metaMap["priority"]; !exists || priority != float64(1) {
		t.Errorf("Expected priority 1, got %v", priority)
	}
}

func TestReadFromNoFrontmatter(t *testing.T) {
	content := `This is just plain content without any frontmatter.

It should still be parsed correctly.`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse plain email: %v", err)
	}

	// Test no frontmatter
	frontmatter := email.FrontMatter()
	if len(frontmatter) != 0 {
		t.Errorf("Should have no frontmatter, got: %s", string(frontmatter))
	}

	// Test content extraction
	emailContent := strings.TrimSpace(string(email.Content()))
	expectedContent := strings.TrimSpace(content)
	if emailContent != expectedContent {
		t.Errorf("Content mismatch.\nExpected: %q\nGot: %q", expectedContent, emailContent)
	}

	// Test metadata is empty
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if metadata != nil {
		t.Errorf("Metadata should be nil for content without frontmatter, got: %v", metadata)
	}

	// Should still be renderable
	if !email.IsRenderable() {
		t.Error("Plain content should be renderable")
	}
}

func TestReadFromHTMLContent(t *testing.T) {
	content := `<html>
<head><title>Test</title></head>
<body>
<p>This is HTML content</p>
</body>
</html>`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse HTML email: %v", err)
	}

	// HTML content should not be renderable (no frontmatter processing)
	if email.IsRenderable() {
		t.Error("HTML content should not be renderable")
	}

	// Should still extract content
	emailContent := strings.TrimSpace(string(email.Content()))
	expectedContent := strings.TrimSpace(content)
	if emailContent != expectedContent {
		t.Errorf("HTML content mismatch.\nExpected: %q\nGot: %q", expectedContent, emailContent)
	}
}

func TestReadFromBOMHandling(t *testing.T) {
	// Content with UTF-8 BOM
	bom := "\ufeff"
	content := bom + `---
subject: "BOM Test"
---

Content after BOM`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse email with BOM: %v", err)
	}

	// Test frontmatter extraction (BOM should be stripped)
	frontmatter := string(email.FrontMatter())
	if strings.Contains(frontmatter, "\ufeff") {
		t.Error("BOM should be stripped from frontmatter")
	}

	if !strings.Contains(frontmatter, `subject: "BOM Test"`) {
		t.Errorf("Frontmatter should contain subject, got: %s", frontmatter)
	}

	// Test metadata
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	metaMap := metadata.(map[string]interface{})
	if subject := metaMap["subject"]; subject != "BOM Test" {
		t.Errorf("Expected subject 'BOM Test', got %v", subject)
	}
}

func TestReadFromWhitespaceHandling(t *testing.T) {
	content := `

   
---
subject: "Whitespace Test"
---


Content with leading whitespace`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse email with whitespace: %v", err)
	}

	// Should still parse frontmatter correctly
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	metaMap := metadata.(map[string]interface{})
	if subject := metaMap["subject"]; subject != "Whitespace Test" {
		t.Errorf("Expected subject 'Whitespace Test', got %v", subject)
	}

	// Content should be preserved
	emailContent := string(email.Content())
	if !strings.Contains(emailContent, "Content with leading whitespace") {
		t.Error("Content should contain the main text")
	}
}

func TestReadFromHTMLCommentsNoFrontmatter(t *testing.T) {
	// HTML comments at the start make content non-renderable (HTML mode)
	content := `<!-- This is a comment at the start -->
---
subject: "HTML Comment Test"
---

Content after HTML comment`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse email with HTML comment: %v", err)
	}

	// HTML content should not be renderable and should not extract frontmatter
	if email.IsRenderable() {
		t.Error("Content starting with HTML comment should not be renderable")
	}

	// Should have no frontmatter extracted
	frontmatter := email.FrontMatter()
	if len(frontmatter) != 0 {
		t.Error("Content starting with HTML comment should not extract frontmatter")
	}

	// Should treat entire content as body
	emailContent := string(email.Content())
	if !strings.Contains(emailContent, "<!-- This is a comment at the start -->") {
		t.Error("Should preserve HTML comment in content")
	}

	if !strings.Contains(emailContent, "subject: \"HTML Comment Test\"") {
		t.Error("Should preserve YAML-like content as regular content (not frontmatter)")
	}

	// Metadata should be nil for HTML content
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	if metadata != nil {
		t.Error("HTML content should have nil metadata")
	}
}

func TestReadFromHTMLCommentsWithFrontmatter(t *testing.T) {
	// Test case where HTML comment is after frontmatter (within content)
	content := `---
subject: "HTML Comment Test"
---

<!-- This is a comment in the content -->
Content after HTML comment`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse email: %v", err)
	}

	// Should be renderable and extract frontmatter correctly
	if !email.IsRenderable() {
		t.Error("Content with frontmatter should be renderable")
	}

	// Should extract frontmatter
	frontmatter := email.FrontMatter()
	if len(frontmatter) == 0 {
		t.Error("Should extract frontmatter")
	}

	// Should parse metadata correctly
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	if metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	metaMap := metadata.(map[string]interface{})
	if subject := metaMap["subject"]; subject != "HTML Comment Test" {
		t.Errorf("Expected subject 'HTML Comment Test', got %v", subject)
	}

	// Content should contain the HTML comment
	emailContent := string(email.Content())
	if !strings.Contains(emailContent, "<!-- This is a comment in the content -->") {
		t.Error("Content should contain HTML comment")
	}
}

func TestReadFromEmptyContent(t *testing.T) {
	reader := strings.NewReader("")
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse empty email: %v", err)
	}

	// Empty content should have no frontmatter
	if len(email.FrontMatter()) != 0 {
		t.Error("Empty content should have no frontmatter")
	}

	// Empty content should have no content
	if len(email.Content()) != 0 {
		t.Error("Empty content should have no content")
	}

	// Empty content is not renderable (no lead character to determine type)
	if email.IsRenderable() {
		t.Error("Empty content should not be renderable")
	}
}

func TestReadFromMalformedYAML(t *testing.T) {
	content := `---
subject: "Test Subject
from: test@example.com
invalid: yaml: content
---

This has malformed YAML frontmatter`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse email: %v", err)
	}

	// Should still extract frontmatter, even if malformed
	frontmatter := email.FrontMatter()
	if len(frontmatter) == 0 {
		t.Error("Should extract frontmatter even if malformed")
	}

	// Metadata parsing should fail
	_, err = email.Metadata()
	if err == nil {
		t.Error("Malformed YAML should cause metadata parsing error")
	}
}

func TestReadFromMultilineContent(t *testing.T) {
	content := `---
subject: "Multiline Test"
description: |
  This is a multiline
  description field
---

Line 1 of content
Line 2 of content

Paragraph 2 here`

	reader := strings.NewReader(content)
	email, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to parse multiline email: %v", err)
	}

	// Test content extraction preserves line breaks
	emailContent := string(email.Content())
	if !strings.Contains(emailContent, "Line 1 of content\nLine 2 of content") {
		t.Error("Should preserve line breaks in content")
	}

	if !strings.Contains(emailContent, "Paragraph 2 here") {
		t.Error("Should preserve all content paragraphs")
	}

	// Test multiline YAML parsing
	metadata, err := email.Metadata()
	if err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	metaMap := metadata.(map[string]interface{})
	if description, exists := metaMap["description"]; !exists {
		t.Error("Should parse multiline description field")
	} else if !strings.Contains(description.(string), "This is a multiline") {
		t.Errorf("Multiline description not parsed correctly: %v", description)
	}
}
