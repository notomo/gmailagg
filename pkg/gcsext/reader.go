package gcsext

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
)

type Reader struct {
	reader *storage.Reader
	client *storage.Client
}

func NewReader(
	ctx context.Context,
	path string,
	baseTransport http.RoundTripper,
) (*Reader, error) {
	client, err := NewClient(ctx, baseTransport, storage.ScopeReadOnly)
	if err != nil {
		return nil, fmt.Errorf("new gcs client: %w", err)
	}

	matches := pathPattern.FindStringSubmatch(path)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	bucket := matches[1]
	object := matches[2]

	reader, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("new reader: %w", err)
	}

	return &Reader{
		reader: reader,
		client: client,
	}, nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *Reader) Close() error {
	defer r.client.Close()
	if err := r.reader.Close(); err != nil {
		return fmt.Errorf("close reader: %w", err)
	}
	return nil
}

func NewReaderByPath(
	ctx context.Context,
	path string,
	baseTransport http.RoundTripper,
) (io.ReadCloser, error) {
	if !isGsutilPath(path) {
		tokenFile, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open file path: %w", err)
		}
		return tokenFile, nil
	}

	reader, err := NewReader(ctx, path, baseTransport)
	if err != nil {
		return nil, fmt.Errorf("new gcs reader: %w", err)
	}
	return reader, nil
}
