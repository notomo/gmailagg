package app

import (
	"fmt"
	"os"

	"github.com/notomo/gmailagg/app/extractor"
	"gopkg.in/yaml.v2"
)

type Influxdb struct {
	ServerURL string `yaml:"serverUrl"`
	Org       string `yaml:"org"`
	Bucket    string `yaml:"bucket"`
}

type Config struct {
	GmailCredentialsPath string                  `yaml:"gmailCredentialsPath"`
	Measurements         []extractor.Measurement `yaml:"measurements"`
	Influxdb             Influxdb                `yaml:"influxdb"`
}

func ReadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	return &config, nil
}
