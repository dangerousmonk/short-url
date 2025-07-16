package config

import (
	"encoding/json"
	"os"
)

func ParseJSONConfig(jsonConfig *Config, fileName string) error {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(file, &jsonConfig); err != nil {
		return err
	}
	return nil
}
