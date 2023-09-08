package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/notomo/gmailagg/pkg/browser"
	"github.com/notomo/gmailagg/pkg/gcsext"
	"github.com/notomo/gmailagg/pkg/gmailext"
)

func Authorize(
	ctx context.Context,
	gmailCredentials string,
	tokenFilePath string,
	opener browser.Opener,
	timeout time.Duration,
	baseTransport http.RoundTripper,
	dryRun bool,
) (retErr error) {
	tokenWriter, err := gcsext.NewWriterByPath(ctx, tokenFilePath, baseTransport, dryRun)
	if err != nil {
		return fmt.Errorf("new token writer: %w", err)
	}
	defer func() {
		if err := tokenWriter.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close token writer: %w", err))
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	token, err := gmailext.Authorize(
		ctx,
		gmailCredentials,
		opener,
		baseTransport,
	)
	if err != nil {
		return fmt.Errorf("gmail authorize: %w", err)
	}

	encorder := json.NewEncoder(tokenWriter)
	encorder.SetIndent("", "  ")
	if err := encorder.Encode(token); err != nil {
		return fmt.Errorf("json encode token: %w", err)
	}

	return nil
}
