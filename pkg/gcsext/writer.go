package gcsext

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"cloud.google.com/go/storage"
)

type Writer struct {
	writer io.WriteCloser
	client *storage.Client
}

var pathPattern = regexp.MustCompile("gs://([^/]+)/(.*)")

func NewWriter(
	ctx context.Context,
	path string,
	baseTransport http.RoundTripper,
) (*Writer, error) {
	client, err := NewClient(ctx, baseTransport)
	if err != nil {
		return nil, fmt.Errorf("new gcs client: %w", err)
	}

	matches := pathPattern.FindStringSubmatch(path)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	bucket := matches[1]
	object := matches[2]

	return &Writer{
		writer: client.Bucket(bucket).Object(object).NewWriter(ctx),
		client: client,
	}, nil
}

func (w *Writer) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *Writer) Close() error {
	defer w.client.Close()
	if err := w.writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}
	return nil
}
