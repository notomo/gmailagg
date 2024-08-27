package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

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

func ReadConfig(
	ctx context.Context,
	path string,
	baseTransport http.RoundTripper,
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
