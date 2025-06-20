package mail

import (
	"path/filepath"
	"testing"

	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
)

func TestParseRecipientsCsv(t *testing.T) {
	aFs := afero.NewMemMapFs()

	// Write and load fake configuration
	cPath, _ := filepath.Abs("./recipients.csv")
	afero.WriteFile(aFs, cPath, []byte(`email,name,extra
jhon.doe@example.com,J Doe,Extra Data
`), 0644)

	appFs := config.Fs{
		Config: &config.AConfig{
			ConfigFile: config.ConfigFile{
				CSV: config.CSVConfig{
					Comma: ",",
				},
			},
		},
		Fs: aFs,
	}

	recipients, err := parseRecipients(&appFs, cPath)
	if err != nil {
		t.Error(err)
	}
	if len(recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(recipients))
	}
}

func TestParseRecipientsYaml(t *testing.T) {
	aFs := afero.NewMemMapFs()

	// Write and load fake configuration
	cPath, _ := filepath.Abs("./recipients.yml")
	afero.WriteFile(aFs, cPath, []byte(`---
- email: jhon.doe@example.com
  name: J Doe
  extra: Extra Data
`), 0644)

	appFs := config.Fs{
		Config: &config.AConfig{
			ConfigFile: config.ConfigFile{
				CSV: config.CSVConfig{
					Comma: ",",
				},
			},
		},
		Fs: aFs,
	}

	recipients, err := parseRecipients(&appFs, cPath)
	if err != nil {
		t.Error(err)
	}
	if len(recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(recipients))
	}
}
