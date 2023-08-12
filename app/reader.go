package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/notomo/gmailagg/pkg/gcsext"
)

func NewTokenReader(
	ctx context.Context,
	tokenFilePath string,
	baseTransport http.RoundTripper,
) (io.ReadCloser, error) {
	if !strings.HasPrefix(tokenFilePath, "gs://") {
		tokenFile, err := os.Open(tokenFilePath)
		if err != nil {
			return nil, fmt.Errorf("open token file path: %w", err)
		}
		return tokenFile, nil
	}

	reader, err := gcsext.NewReader(ctx, tokenFilePath, baseTransport)
	if err != nil {
		return nil, fmt.Errorf("new gcs reader: %w", err)
	}
	return reader, nil
}
