package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/notomo/gmailagg/app/extractor"
)

type Config struct {
	Measurements []extractor.Measurement `json:"measurements"`
}

func ReadConfig(
	path string,
) (_ *Config, retErr error) {
	reader, err := createTokenReader(path)
	if err != nil {
		return nil, fmt.Errorf("new config reader: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close config reader: %w", err))
		}
	}()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &config, nil
}
