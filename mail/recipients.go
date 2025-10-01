package mail

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/rykov/paperboy/config"
)

// unmarshalCsvRecipients parses CSV-formatted recipient data into a slice of maps
// Each map represents one recipient with CSV headers as keys
func unmarshalCsvRecipients(cfg *config.CSVConfig, raw []byte) ([]map[string]any, error) {
	csvReader := csv.NewReader(bytes.NewReader(raw))

	// Validate and apply custom separator
	if l := len(cfg.Separator); l > 1 {
		return nil, errors.New("multi-character CSV separator not supported")
	} else if l == 1 {
		csvReader.Comma = []rune(cfg.Separator)[0]
	}

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	data := make([]map[string]any, 0)

	// Map CSV rows to recipient dictionaries
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("CSV parse error: %w", err)
		}

		rec := make(map[string]any)
		for i, h := range header {
			rec[h] = record[i]
		}
		data = append(data, rec)
	}

	return data, nil
}
