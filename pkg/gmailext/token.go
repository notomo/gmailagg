package gmailext

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/notomo/gmailagg/pkg/browser"
	"github.com/notomo/gmailagg/pkg/httpext"
	"golang.org/x/oauth2"
)

type Browser interface {
	Open(url string) error
}

func Authorize(
	ctx context.Context,
	credentialsJsonPath string,
	opener browser.Opener,
) (token *oauth2.Token, retErr error) {
	config, err := getOauth2Config(credentialsJsonPath)
	if err != nil {
		return nil, fmt.Errorf("get oauth2 config: %w", err)
	}

	authCodeReceiver := make(chan string)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		// TODO: error handling

		var authCode string
		query := req.URL.Query()
		code, ok := query["code"]
		if ok && len(code) > 0 {
			authCode = code[0]
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`ok`)); err != nil {
			retErr = errors.Join(retErr, err)
		}

		// TODO: redirect

		authCodeReceiver <- authCode
	})

	url, server, listener, err := httpext.NewServer(mux)
	if err != nil {
		return nil, fmt.Errorf("new server: %w", err)
	}
	go func() {
		if err := server.Serve(listener); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	config.RedirectURL = url + "/callback"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	if err := opener.Open(ctx, authURL); err != nil {
		return nil, fmt.Errorf("browser open: %w", err)
	}

	authCode := <-authCodeReceiver
	token, err = config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("exchange: %w", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		return nil, fmt.Errorf("server shutdown: %w", err)
	}

	return token, nil
}
