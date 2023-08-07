package gmailext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	ghttp "google.golang.org/api/transport/http"
)

func NewService(
	ctx context.Context,
	tokenFilePath string,
	baseTransport http.RoundTripper,
) (*gmail.Service, error) {
	config, err := getOauth2Config(ctx)
	if err != nil {
		return nil, fmt.Errorf("get oauth2 config: %w", err)
	}

	var token oauth2.Token
	{
		f, err := os.Open(tokenFilePath)
		if err != nil {
			return nil, fmt.Errorf("open token file path: %w", err)
		}
		defer f.Close()

		if err := json.NewDecoder(f).Decode(&token); err != nil {
			return nil, fmt.Errorf("json decode token: %w", err)
		}
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Timeout:   20 * time.Second,
		Transport: baseTransport,
	})
	tokenSource := config.TokenSource(ctx, &token)
	transport, err := ghttp.NewTransport(ctx, baseTransport, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("new transport: %w", err)
	}

	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: transport,
	}
	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("new service: %w", err)
	}

	return service, nil
}

const gmailScope = gmail.GmailReadonlyScope

func getOauth2Config(ctx context.Context) (*oauth2.Config, error) {
	credentials, err := google.FindDefaultCredentials(ctx, gmailScope)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(credentials.JSON, gmailScope)
	if err != nil {
		return nil, fmt.Errorf("config from json: %w", err)
	}
	return config, nil
}
