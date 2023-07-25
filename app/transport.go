package app

import (
	"net/http"

	"github.com/henvic/httpretty"
	"github.com/notomo/httpwriter"
)

func LogTransport(logDirPath string, baseTransport http.RoundTripper) http.RoundTripper {
	if logDirPath == "" {
		return baseTransport
	}

	logger := &httpretty.Logger{
		Time:            true,
		TLS:             false,
		RequestHeader:   true,
		RequestBody:     true,
		ResponseHeader:  true,
		ResponseBody:    true,
		MaxResponseBody: 1000000,
		Formatters:      []httpretty.Formatter{&httpretty.JSONFormatter{}},
	}
	return &httpwriter.Transport{
		Transport: logger.RoundTripper(baseTransport),
		GetWriter: httpwriter.MustDirectoryWriter(
			&httpwriter.Directory{Path: logDirPath},
		),
		SetWriter: logger.SetOutput,
	}
}
