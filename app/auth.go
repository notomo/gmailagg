package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/notomo/gmailagg/pkg/browser"
	"github.com/notomo/gmailagg/pkg/gmailext"
)

func Authorize(
	ctx context.Context,
	gmailCredentials string,
	tokenFilePath string,
	opener browser.Opener,
	baseTransport http.RoundTripper,
) (retErr error) {
	tokenWriter, err := NewTokenWriter(ctx, tokenFilePath, baseTransport)
	if err != nil {
		return fmt.Errorf("new token writer: %w", err)
	}
	defer func() {
		if err := tokenWriter.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close token writer: %w", err))
		}
	}()

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
