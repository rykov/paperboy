package mail

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
)

func TestVerifyDKIMForMail(t *testing.T) {
	// Test with invalid mail data
	t.Run("invalid mail data", func(t *testing.T) {
		err := verifyDKIMForMail([]byte("invalid mail data"))
		if err == nil {
			t.Error("Expected error for invalid mail data")
		}
		if !strings.Contains(err.Error(), "DKIM verification failed") {
			t.Errorf("Expected 'DKIM verification failed' in error, got: %v", err)
		}
	})

	// Test with mail without DKIM signatures
	t.Run("mail without DKIM", func(t *testing.T) {
		mailWithoutDKIM := []byte(`From: test@example.com
To: recipient@example.com
Subject: Test Email

This is a test email without DKIM signature.`)

		err := verifyDKIMForMail(mailWithoutDKIM)
		if err == nil {
			t.Error("Expected error for mail without DKIM")
		}
		if !strings.Contains(err.Error(), "no DKIM signatures found") {
			t.Errorf("Expected 'no DKIM signatures found' in error, got: %v", err)
		}
	})
}

// Helper function to create test config
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
	t.Run("no DKIM configuration - should succeed", func(t *testing.T) {
		cfg := newTestConfig(map[string]interface{}{}) // Empty DKIM config

		// Since DKIM is optional, this should fail on campaign/email rendering
		err := VerifyCampaign(cfg, "test-campaign", "test-list")
		if err == nil {
			t.Error("Expected error for campaign with no content")
		}

		// Could fail either on campaign loading or no emails rendered
		if !strings.Contains(err.Error(), "failed to load campaign") &&
			!strings.Contains(err.Error(), "no emails were rendered") {
			t.Errorf("Expected campaign or email rendering error, got: %v", err)
		}
	})

	t.Run("invalid campaign with DKIM config", func(t *testing.T) {
		cfg := newTestConfig(map[string]interface{}{
			"domain": "example.com", // At least one DKIM config
		})

		err := VerifyCampaign(cfg, "nonexistent-campaign", "nonexistent-list")
		if err == nil {
			t.Error("Expected error for nonexistent campaign")
		}

		expectedMsg := "failed to load campaign"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error to contain %q, got: %v", expectedMsg, err)
		}
	})
}

func TestVerifyCampaignErrorMessages(t *testing.T) {
	testCases := []struct {
		name          string
		setupDKIM     bool
		tmplFile      string
		recipientFile string
		expectedError string
	}{
		{
			name:          "invalid template without DKIM",
			setupDKIM:     false,
			tmplFile:      "test",
			recipientFile: "test",
			expectedError: "no emails were rendered",
		},
		{
			name:          "invalid template with DKIM",
			setupDKIM:     true,
			tmplFile:      "test",
			recipientFile: "test",
			expectedError: "failed to load campaign",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var dkimConfig map[string]interface{}
			if tc.setupDKIM {
				dkimConfig = map[string]interface{}{
					"domain": "example.com",
				}
			} else {
				dkimConfig = map[string]interface{}{}
			}

			cfg := newTestConfig(dkimConfig)
			err := VerifyCampaign(cfg, tc.tmplFile, tc.recipientFile)

			if err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("Expected error to contain %q, got: %v", tc.expectedError, err)
			}
		})
	}
}

func TestVerifyCampaignValidation(t *testing.T) {
	t.Run("loads campaign first regardless of DKIM config", func(t *testing.T) {
		cfg := newTestConfig(map[string]interface{}{}) // No DKIM

		// Should fail on campaign loading since DKIM is now optional
		err := VerifyCampaign(cfg, "", "")
		if err == nil {
			t.Error("Expected error for invalid campaign")
		}

		// Could fail on campaign loading or no emails rendered
		if !strings.Contains(err.Error(), "failed to load campaign") &&
			!strings.Contains(err.Error(), "no emails were rendered") {
			t.Errorf("Expected campaign or email rendering error, got: %v", err)
		}
	})

	t.Run("loads campaign first with DKIM config", func(t *testing.T) {
		cfg := newTestConfig(map[string]interface{}{
			"domain": "example.com", // Valid DKIM config
		})

		err := VerifyCampaign(cfg, "invalid-campaign", "invalid-list")
		if err == nil {
			t.Error("Expected error for invalid campaign")
		}

		// Should fail on campaign loading
		if !strings.Contains(err.Error(), "failed to load campaign") {
			t.Errorf("Expected campaign loading error, got: %v", err)
		}
	})
}

// Helper function to create mock verification results
func createMockDKIMError(domain string, err error) error {
	return errors.New("mock DKIM error for testing")
}
