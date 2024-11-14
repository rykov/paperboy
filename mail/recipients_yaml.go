package mail

import (
	"github.com/ghodss/yaml"
	"github.com/rykov/paperboy/config"
)

func unmarshalYamlRecipients(appFs *config.Fs, raw []byte) ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return data, nil
}
