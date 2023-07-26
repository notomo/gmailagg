package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/notomo/gmailagg/pkg/browser"
	"github.com/notomo/gmailagg/pkg/gmailext"
)

func Authorize(
	ctx context.Context,
	credentialsJsonPath string,
	tokenFilePath string,
	opener browser.Opener,
	baseTransport http.RoundTripper,
) error {
	token, err := gmailext.Authorize(
		ctx,
		credentialsJsonPath,
		opener,
		baseTransport,
	)
	if err != nil {
		return fmt.Errorf("gmail authorize: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(tokenFilePath), 0700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	tokenFile, err := os.OpenFile(tokenFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open token file: %w", err)
	}
	defer tokenFile.Close()

	encorder := json.NewEncoder(tokenFile)
	encorder.SetIndent("", "  ")
	if err := encorder.Encode(token); err != nil {
		return fmt.Errorf("json encode token: %w", err)
	}

	return nil
}

func TokenFilePath() string {
	return filepath.Join(xdg.ConfigHome, "gmailagg/token.json")
}
