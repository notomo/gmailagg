package gmailext

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/notomo/gmailagg/pkg/browser"
	"github.com/notomo/gmailagg/pkg/httpext"
	"golang.org/x/oauth2"
)

func Authorize(
	ctx context.Context,
	gmailCredentials string,
	opener browser.Opener,
	baseTransport http.RoundTripper,
) (token *oauth2.Token, retErr error) {
	config, err := getOauth2Config(ctx, gmailCredentials)
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

	serveErr := make(chan error)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	config.RedirectURL = url + "/callback"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	if err := opener.Open(ctx, authURL); err != nil {
		return nil, fmt.Errorf("browser open: %w", err)
	}

	var authCode string
	select {
	case authCode = <-authCodeReceiver:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Timeout:   20 * time.Second,
		Transport: baseTransport,
	})
	token, err = config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("exchange: %w", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		return nil, fmt.Errorf("server shutdown: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return nil, err
	}

	return token, nil
}
