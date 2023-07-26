package gmailtest

import (
	"context"
	"net/http"
	"net/url"
)

type Opener struct {
	AuthCode string
}

func (o *Opener) Open(ctx context.Context, u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return err
	}

	query := parsed.Query()
	parsedRedirectURL, err := url.Parse(query.Get("redirect_uri"))
	if err != nil {
		return err
	}

	newQuery := parsedRedirectURL.Query()
	newQuery.Set("code", o.AuthCode)
	parsedRedirectURL.RawQuery = newQuery.Encode()

	go func() {
		if _, err := http.Get(parsedRedirectURL.String()); err != nil {
			panic(err)
		}
	}()

	return nil
}
