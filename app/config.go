package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/notomo/gmailagg/app/extractor"
)

type Influxdb struct {
	ServerURL string `json:"serverUrl"`
	Org       string `json:"org"`
	Bucket    string `json:"bucket"`
}

type Config struct {
	Measurements []extractor.Measurement `json:"measurements"`
	Influxdb     Influxdb                `json:"influxdb"`
}

func ReadConfig(path string, s string) (*Config, error) {
	var content []byte
	if path == "" {
		content = []byte(s)
	} else {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
		content = b
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &config, nil
}
