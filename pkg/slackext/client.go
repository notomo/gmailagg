package slackext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	webhookURL    string
	baseTransport http.RoundTripper
}

func New(
	webhookURL string,
	baseTransport http.RoundTripper,
) *Client {
	return &Client{
		webhookURL:    webhookURL,
		baseTransport: baseTransport,
	}
}

func (c *Client) Open(ctx context.Context, u string) error {
	b, err := json.Marshal(map[string]string{"text": u})
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	buf := bytes.NewBuffer(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, buf)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	httpClient := &http.Client{
		Timeout:   20 * time.Second,
		Transport: c.baseTransport,
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer res.Body.Close()

	return nil
}
