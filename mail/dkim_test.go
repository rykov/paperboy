package mail

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"strings"
	"testing"

	"github.com/go-gomail/gomail"
	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
)

// Generate a test RSA private key for DKIM testing
func generateTestPrivateKey() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return privateKeyPEM, nil
}

// Mock SendCloser for testing
type mockSendCloser struct {
	messages []mockMessage
	closed   bool
}

type mockMessage struct {
	from string
	to   []string
	body string
}

func (m *mockSendCloser) Send(from string, to []string, msg io.WriterTo) error {
	var buf bytes.Buffer
	_, err := msg.WriteTo(&buf)
	if err != nil {
		return err
	}

	m.messages = append(m.messages, mockMessage{
		from: from,
		to:   to,
		body: buf.String(),
	})
	return nil
}

func (m *mockSendCloser) Close() error {
	m.closed = true
	return nil
}

func TestDKIMSendCloserSuccess(t *testing.T) {
	// Generate test private key
	privateKeyPEM, err := generateTestPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate test private key: %v", err)
	}

	// Setup virtual filesystem with DKIM key
	memFs := afero.NewMemMapFs()
	keyPath := "/dkim/private.key"
	afero.WriteFile(memFs, keyPath, privateKeyPEM, 0600)

	cfg, err := config.LoadConfigFs(t.Context(), memFs)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// DKIM configuration
	dkimConfig := map[string]interface{}{
		"keyfile":  keyPath,
		"domain":   "example.com",
		"selector": "default",
	}

	// Create mock send closer
	mockSC := &mockSendCloser{}

	// Create DKIM-enabled sender
	dkimSC, err := SendCloserWithDKIM(cfg.AppFs, mockSC, dkimConfig)
	if err != nil {
		t.Fatalf("Failed to create DKIM send closer: %v", err)
	}

	// Create test message
	msg := gomail.NewMessage()
	msg.SetHeader("From", "test@example.com")
	msg.SetHeader("To", "recipient@example.com")
	msg.SetHeader("Subject", "Test Email")
	msg.SetBody("text/plain", "This is a test email for DKIM signing")

	// Send message through DKIM sender
	err = dkimSC.Send("test@example.com", []string{"recipient@example.com"}, msg)
	if err != nil {
		t.Fatalf("Failed to send message through DKIM sender: %v", err)
	}

	// Verify message was sent
	if len(mockSC.messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(mockSC.messages))
	}

	message := mockSC.messages[0]

	// Verify DKIM signature is present
	if !strings.Contains(message.body, "DKIM-Signature:") {
		t.Error("Message should contain DKIM-Signature header")
	}

	// Verify DKIM signature contains expected parameters
	if !strings.Contains(message.body, "d=example.com") {
		t.Error("DKIM signature should contain domain (d=example.com)")
	}

	if !strings.Contains(message.body, "s=default") {
		t.Error("DKIM signature should contain selector (s=default)")
	}

	// Verify Close() works
	err = dkimSC.Close()
	if err != nil {
		t.Errorf("Failed to close DKIM sender: %v", err)
	}

	if !mockSC.closed {
		t.Error("Underlying sender should be closed")
	}
}

func TestDKIMSendCloserMissingKeyFile(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	dkimConfig := map[string]interface{}{
		"domain":   "example.com",
		"selector": "default",
		// Missing keyfile
	}

	mockSC := &mockSendCloser{}
	_, err := SendCloserWithDKIM(cfg.AppFs, mockSC, dkimConfig)

	if err == nil {
		t.Error("Should return error when keyfile is missing")
	}

	if !strings.Contains(err.Error(), "DKIM requires a keyFile") {
		t.Errorf("Error should mention missing keyFile, got: %v", err)
	}
}

func TestDKIMSendCloserInvalidKeyFile(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	dkimConfig := map[string]interface{}{
		"keyfile":  "/nonexistent/key.pem",
		"domain":   "example.com",
		"selector": "default",
	}

	mockSC := &mockSendCloser{}
	_, err := SendCloserWithDKIM(cfg.AppFs, mockSC, dkimConfig)

	if err == nil {
		t.Error("Should return error when keyfile doesn't exist")
	}
}

func TestDKIMSendCloserOptionalParameters(t *testing.T) {
	// Generate test private key
	privateKeyPEM, err := generateTestPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate test private key: %v", err)
	}

	// Setup virtual filesystem with DKIM key
	memFs := afero.NewMemMapFs()
	keyPath := "/dkim/private.key"
	afero.WriteFile(memFs, keyPath, privateKeyPEM, 0600)

	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	// DKIM configuration with optional parameters
	dkimConfig := map[string]interface{}{
		"keyfile":           keyPath,
		"domain":            "example.com",
		"selector":          "default",
		"signatureexpirein": 3600,
		"canonicalization":  "relaxed/simple",
	}

	mockSC := &mockSendCloser{}
	dkimSC, err := SendCloserWithDKIM(cfg.AppFs, mockSC, dkimConfig)
	if err != nil {
		t.Fatalf("Failed to create DKIM send closer with optional params: %v", err)
	}

	// Verify it was created successfully
	if dkimSC == nil {
		t.Error("DKIM send closer should not be nil")
	}
}

func TestDKIMMessageWriteTo(t *testing.T) {
	// Generate test private key
	privateKeyPEM, err := generateTestPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate test private key: %v", err)
	}

	// Setup virtual filesystem with DKIM key
	memFs := afero.NewMemMapFs()
	keyPath := "/dkim/private.key"
	afero.WriteFile(memFs, keyPath, privateKeyPEM, 0600)

	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	dkimConfig := map[string]interface{}{
		"keyfile":  keyPath,
		"domain":   "example.com",
		"selector": "default",
	}

	mockSC := &mockSendCloser{}
	dkimSC, err := SendCloserWithDKIM(cfg.AppFs, mockSC, dkimConfig)
	if err != nil {
		t.Fatalf("Failed to create DKIM send closer: %v", err)
	}

	// Create test message
	msg := gomail.NewMessage()
	msg.SetHeader("From", "test@example.com")
	msg.SetHeader("To", "recipient@example.com")
	msg.SetHeader("Subject", "Test Email")
	msg.SetBody("text/plain", "Test content")

	// Test direct WriteTo functionality
	var buf bytes.Buffer
	dkimWrapper := dkimSC.(*dkimSendCloser)
	dkimMsg := dkimMessage{
		options: dkimWrapper.Options,
		msg:     msg,
	}

	n, err := dkimMsg.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write DKIM message: %v", err)
	}

	if n == 0 {
		t.Error("Should write some bytes")
	}

	output := buf.String()
	if !strings.Contains(output, "DKIM-Signature:") {
		t.Error("Output should contain DKIM-Signature header")
	}

	// Verify original message headers are preserved
	if !strings.Contains(output, "From: test@example.com") {
		t.Error("Original From header should be preserved")
	}

	if !strings.Contains(output, "Subject: Test Email") {
		t.Error("Original Subject header should be preserved")
	}
}
