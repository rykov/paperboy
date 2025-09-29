package mail

import (
	"context"
	"strings"
	"testing"

	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
)

func TestVerifyDKIMForMail(t *testing.T) {
	tests := []struct {
		name     string
		mailData []byte
		expected string
	}{
		{
			name:     "invalid mail data",
			mailData: []byte("invalid mail data"),
			expected: "DKIM verification failed",
		},
		{
			name: "mail without DKIM",
			mailData: []byte(`From: test@example.com
To: recipient@example.com
Subject: Test Email

This is a test email without DKIM signature.`),
			expected: "no DKIM signatures found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyDKIMForMail(tt.mailData)
			assertError(t, err, tt.expected)
		})
	}
}

// Helper functions
func assertError(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected error containing %q but got none", expectedMsg)
		return
	}
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain %q, got: %v", expectedMsg, err)
	}
}

func assertCampaignError(t *testing.T, err error, expectedMsgs ...string) {
	t.Helper()
	if err == nil {
		t.Error("Expected error but got none")
		return
	}
	for _, msg := range expectedMsgs {
		if strings.Contains(err.Error(), msg) {
			return
		}
	}
	t.Errorf("Expected error to contain one of %v, got: %v", expectedMsgs, err)
}

func newTestConfig(dkimConfig map[string]interface{}) *config.AConfig {
	fs := afero.NewMemMapFs()

	cfg := &config.AConfig{
		ConfigFile: config.ConfigFile{
			DKIM:       dkimConfig,
			ContentDir: "content",
			ListDir:    "lists",
		},
		Context: context.Background(),
	}

	// Set up the AppFs with proper back-reference
	cfg.AppFs = &config.Fs{
		Config: cfg,
		Fs:     fs,
	}

	return cfg
}

func TestVerifyCampaign(t *testing.T) {
	tests := []struct {
		name          string
		dkimConfig    map[string]interface{}
		tmplFile      string
		recipientFile string
		expectedMsgs  []string
	}{
		{
			name:          "no DKIM configuration",
			dkimConfig:    map[string]interface{}{},
			tmplFile:      "test-campaign",
			recipientFile: "test-list",
			expectedMsgs:  []string{"failed to load campaign", "no emails were rendered"},
		},
		{
			name:          "invalid campaign with DKIM config",
			dkimConfig:    map[string]interface{}{"domain": "example.com"},
			tmplFile:      "nonexistent-campaign",
			recipientFile: "nonexistent-list",
			expectedMsgs:  []string{"failed to load campaign"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newTestConfig(tt.dkimConfig)
			err := VerifyCampaign(cfg, tt.tmplFile, tt.recipientFile)
			assertCampaignError(t, err, tt.expectedMsgs...)
		})
	}
}

// TestVerifyCampaignErrorMessages is covered by TestVerifyCampaign

// TestVerifyCampaignValidation is covered by TestVerifyCampaign

func TestCheckDuplicateEmails(t *testing.T) {
	tests := []struct {
		name        string
		recipients  []*ctxRecipient
		expectError bool
		errorMsg    string
	}{
		{
			name: "no duplicates",
			recipients: []*ctxRecipient{
				{"email": "user1@example.com", "name": "User 1"},
				{"email": "user2@example.com", "name": "User 2"},
				{"email": "user3@example.com", "name": "User 3"},
			},
			expectError: false,
		},
		{
			name: "exact duplicate emails",
			recipients: []*ctxRecipient{
				{"email": "user1@example.com", "name": "User 1"},
				{"email": "user2@example.com", "name": "User 2"},
				{"email": "user1@example.com", "name": "User 1 Duplicate"},
			},
			expectError: true,
			errorMsg:    "duplicate email address found: \"user1@example.com\" (first seen at index 0, duplicate at index 2)",
		},
		{
			name: "case-insensitive duplicates",
			recipients: []*ctxRecipient{
				{"email": "user1@example.com", "name": "User 1"},
				{"email": "USER1@EXAMPLE.COM", "name": "User 1 Uppercase"},
			},
			expectError: true,
			errorMsg:    "duplicate email address found: \"USER1@EXAMPLE.COM\" (first seen at index 0, duplicate at index 1)",
		},
		{
			name: "whitespace normalized duplicates",
			recipients: []*ctxRecipient{
				{"email": "user1@example.com", "name": "User 1"},
				{"email": " user1@example.com ", "name": "User 1 With Spaces"},
			},
			expectError: true,
			errorMsg:    "duplicate email address found: \" user1@example.com \" (first seen at index 0, duplicate at index 1)",
		},
		{
			name: "empty email address",
			recipients: []*ctxRecipient{
				{"email": "user1@example.com", "name": "User 1"},
				{"email": "", "name": "User 2"},
			},
			expectError: true,
			errorMsg:    "recipient at index 1 has empty email address",
		},
		{
			name: "whitespace-only email address",
			recipients: []*ctxRecipient{
				{"email": "user1@example.com", "name": "User 1"},
				{"email": "   ", "name": "User 2"},
			},
			expectError: true,
			errorMsg:    "recipient at index 1 has empty email address",
		},
		{
			name: "mixed case and spacing normalization",
			recipients: []*ctxRecipient{
				{"email": "User1@Example.Com", "name": "User 1"},
				{"email": " user1@example.com ", "name": "User 1 Normalized"},
				{"email": "USER2@EXAMPLE.COM", "name": "User 2"},
			},
			expectError: true,
			errorMsg:    "duplicate email address found: \" user1@example.com \" (first seen at index 0, duplicate at index 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkDuplicateEmails(tt.recipients)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got: %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
