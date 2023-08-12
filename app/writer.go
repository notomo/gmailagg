package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/notomo/gmailagg/pkg/gcsext"
)

func NewTokenWriter(
	ctx context.Context,
	tokenFilePath string,
	baseTransport http.RoundTripper,
) (io.WriteCloser, error) {
	if !strings.HasPrefix(tokenFilePath, "gs://") {
		if err := os.MkdirAll(filepath.Dir(tokenFilePath), 0700); err != nil {
			return nil, fmt.Errorf("mkdir: %w", err)
		}

		tokenFile, err := os.OpenFile(tokenFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return nil, fmt.Errorf("open token file: %w", err)
		}

		return tokenFile, nil
	}

	writer, err := gcsext.NewWriter(ctx, tokenFilePath, baseTransport)
	if err != nil {
		return nil, fmt.Errorf("new gcs writer: %w", err)
	}
	return writer, nil
}
