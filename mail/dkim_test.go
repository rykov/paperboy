package mail

import (
	"github.com/emersion/go-msgauth/dkim"
	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
	"github.com/wneessen/go-mail"

	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"strings"
	"testing"
)

// Generate a test RSA private key for DKIM testing
func generateTestPrivateKey() (*rsa.PrivateKey, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(key),
		Type:  "RSA PRIVATE KEY",
	})

	return key, keyPEM, nil
}

func TestDKIMMiddlewareSuccess(t *testing.T) {
	// Generate test private key
	rsaKey, privateKeyPEM, err := generateTestPrivateKey()
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
	cfg.DKIM = map[string]interface{}{
		"keyfile":  keyPath,
		"domain":   "example.com",
		"selector": "default",
	}

	// Create DKIM middleware
	middleware, err := DKIMMiddleware(cfg)
	if err != nil {
		t.Fatalf("Failed to create DKIM middleware: %v", err)
	}

	if middleware == nil {
		t.Error("DKIM middleware should not be nil")
	}

	// Create test message with middleware
	msg := mail.NewMsg(mail.WithMiddleware(middleware))
	if err := msg.From("test@example.com"); err != nil {
		t.Fatalf("Failed to set From: %v", err)
	}
	if err := msg.To("recipient@example.com"); err != nil {
		t.Fatalf("Failed to set To: %v", err)
	}
	msg.Subject("Test Email")
	msg.SetBodyString(mail.TypeTextPlain, "This is a test email for DKIM signing")

	// Write message and verify DKIM signature is present
	var buf bytes.Buffer
	if _, err := msg.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	msgContent := buf.String()
	if !strings.Contains(msgContent, "DKIM-Signature:") {
		t.Error("Message should contain DKIM-Signature header")
	}

	// Verify DKIM signature contains expected parameters
	if !strings.Contains(msgContent, "d=example.com") {
		t.Error("DKIM signature should contain domain (d=example.com)")
	}

	if !strings.Contains(msgContent, "s=default") {
		t.Error("DKIM signature should contain selector (s=default)")
	}

	verifyEmailWithDKIM(t, &buf, rsaKey)
}

func TestDKIMMiddlewareMissingKeyFile(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	cfg.DKIM = map[string]interface{}{
		"domain":   "example.com",
		"selector": "default",
		// Missing keyfile
	}

	_, err := DKIMMiddleware(cfg)

	if err == nil {
		t.Error("Should return error when keyfile is missing")
	}

	if !strings.Contains(err.Error(), "DKIM requires a keyFile") {
		t.Errorf("Error should mention missing keyFile, got: %v", err)
	}
}

func TestDKIMMiddlewareInvalidKeyFile(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	cfg.DKIM = map[string]interface{}{
		"keyfile":  "/nonexistent/key.pem",
		"domain":   "example.com",
		"selector": "default",
	}

	_, err := DKIMMiddleware(cfg)

	if err == nil {
		t.Error("Should return error when keyfile doesn't exist")
	}
}

func TestDKIMMiddlewareOptionalParameters(t *testing.T) {
	// Generate test private key
	_, privateKeyPEM, err := generateTestPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate test private key: %v", err)
	}

	// Setup virtual filesystem with DKIM key
	memFs := afero.NewMemMapFs()
	keyPath := "/dkim/private.key"
	afero.WriteFile(memFs, keyPath, privateKeyPEM, 0600)

	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	// DKIM configuration with optional parameters
	cfg.DKIM = map[string]interface{}{
		"keyfile":           keyPath,
		"domain":            "example.com",
		"selector":          "default",
		"signatureexpirein": 3600,
		"canonicalization":  "relaxed/simple",
	}

	middleware, err := DKIMMiddleware(cfg)
	if err != nil {
		t.Fatalf("Failed to create DKIM middleware with optional params: %v", err)
	}

	// Verify it was created successfully
	if middleware == nil {
		t.Error("DKIM middleware should not be nil")
	}
}

func TestMsgOptionsWithDKIM(t *testing.T) {
	// Generate test private key
	rsaKey, privateKeyPEM, err := generateTestPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate test private key: %v", err)
	}

	// Setup virtual filesystem with DKIM key
	memFs := afero.NewMemMapFs()
	keyPath := "/dkim/private.key"
	afero.WriteFile(memFs, keyPath, privateKeyPEM, 0600)

	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	cfg.DKIM = map[string]interface{}{
		"keyfile":  keyPath,
		"domain":   "example.com",
		"selector": "default",
	}

	// Test msgOptions function
	opts, err := msgOptions(cfg)
	if err != nil {
		t.Fatalf("Failed to create message options: %v", err)
	}

	// Should have middleware option
	if len(opts) == 0 {
		t.Error("Should have at least one message option for DKIM")
	}

	// Create test message with the options
	msg := mail.NewMsg(opts...)
	if err := msg.From("test@example.com"); err != nil {
		t.Fatalf("Failed to set From: %v", err)
	}
	if err := msg.To("recipient@example.com"); err != nil {
		t.Fatalf("Failed to set To: %v", err)
	}
	msg.Subject("Test Email")
	msg.SetBodyString(mail.TypeTextPlain, "Test content")

	// Write message and verify DKIM signature
	var buf bytes.Buffer
	if _, err := msg.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "DKIM-Signature:") {
		t.Error("Output should contain DKIM-Signature header")
	}

	// Verify original message headers are preserved
	if !strings.Contains(output, "test@example.com") {
		t.Error("Original From header should be preserved")
	}

	if !strings.Contains(output, "Subject: Test Email") {
		t.Error("Original Subject header should be preserved")
	}

	verifyEmailWithDKIM(t, &buf, rsaKey)
}

func TestMsgOptionsWithoutDKIM(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := config.LoadConfigFs(t.Context(), memFs)

	// No DKIM configuration
	cfg.DKIM = map[string]interface{}{}

	// Test msgOptions function
	opts, err := msgOptions(cfg)
	if err != nil {
		t.Fatalf("Failed to create message options: %v", err)
	}

	// Should have no options
	if len(opts) != 0 {
		t.Error("Should have no message options when DKIM is not configured")
	}
}

// Decode and verify DKIM signature for reader of incoming email
func verifyEmailWithDKIM(t *testing.T, r io.Reader, rsaKey *rsa.PrivateKey) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKeyBytes)
	verifications, err := dkim.VerifyWithOptions(r, &dkim.VerifyOptions{
		LookupTXT: func(domain string) ([]string, error) {
			if domain == "default._domainkey.example.com" {
				return []string{fmt.Sprintf("v=DKIM1; k=rsa; p=%s", pubKeyB64)}, nil
			}
			return nil, fmt.Errorf("DNS record not found for domain: %s", domain)
		},
	})

	if err != nil {
		t.Errorf("DKIM Verification error: %s", err)
	} else if len(verifications) == 0 {
		t.Errorf("DKIM Verification missing")
	}

	for _, v := range verifications {
		if v.Err == nil {
			t.Logf("Valid DKIM signature for domain: %s", v.Domain)
		} else {
			t.Errorf("Invalid DKIM signature for domain %s: %s", v.Domain, v.Err)
		}
	}
}
