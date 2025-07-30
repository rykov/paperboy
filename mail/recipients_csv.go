package mail

import (
	"bytes"
	"encoding/csv"
	"io"

	"github.com/rykov/paperboy/config"
)

func unmarshalCsvRecipients(appFs *config.Fs, raw []byte) ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	// CSV
	csvReader := newCSVReader(appFs.Config.ConfigFile.CSV, bytes.NewReader(raw))

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	// Read CSV line by line
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// Create a map for each record
		rec := make(map[string]interface{})
		for i, h := range header {
			rec[h] = record[i]
		}

		data = append(data, rec)
	}

	return data, nil
}

func newCSVReader(cfg config.CSVConfig, r io.Reader) *csv.Reader {
	csvReader := csv.NewReader(r)
	// Deal with separator
	if len(cfg.Comma) > 0 {
		csvReader.Comma = []rune(cfg.Comma)[0]
	}
	return csvReader
}
