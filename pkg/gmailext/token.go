package gmailext

import (
	"context"
	"fmt"
	"io"

	"golang.org/x/oauth2"
)

func Authorize(
	ctx context.Context,
	credentialsJsonPath string,
	messageWriter io.Writer,
	inputReader io.Reader,
) (*oauth2.Token, error) {
	config, err := getOauth2Config(credentialsJsonPath)
	if err != nil {
		return nil, fmt.Errorf("get oauth2 config: %w", err)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	message := fmt.Sprintf(`Go to the following link in your browser then type the authorization code:
%s
`, authURL)
	if _, err := messageWriter.Write([]byte(message)); err != nil {
		return nil, fmt.Errorf("write message: %w", err)
	}

	var authCode string
	if _, err := fmt.Fscan(inputReader, &authCode); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("exchange: %w", err)
	}

	return token, nil
}
