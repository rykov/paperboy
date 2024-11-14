package mail

import (
	"testing"

	"github.com/rykov/paperboy/config"
)

func TestUnmarshalCsv(t *testing.T) {
	content := []byte(`email,name,extra
jhon.doe@example.com,J Doe,Extra Data
`)

	appFs := config.Fs{
		Config: &config.AConfig{
			ConfigFile: config.ConfigFile{
				CSV: config.CSVConfig{
					Comma: ",",
				},
			},
		},
	}

	recipients, err := unmarshalCsvRecipients(&appFs, content)
	if err != nil {
		t.Error(err)
	}
	if len(recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(recipients))
	}
}
