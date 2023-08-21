package gcsext

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
)

type Writer struct {
	writer *storage.Writer
	client *storage.Client
}

var pathPattern = regexp.MustCompile("gs://([^/]+)/(.*)")

func NewWriter(
	ctx context.Context,
	path string,
	baseTransport http.RoundTripper,
) (*Writer, error) {
	client, err := NewClient(ctx, baseTransport, storage.ScopeReadWrite)
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

func isGsutilPath(path string) bool {
	return strings.HasPrefix(path, "gs://")
}

type discard struct {
}

func (d *discard) Write(p []byte) (n int, err error) {
	return io.Discard.Write(p)
}

func (d *discard) Close() error {
	return nil
}

func NewWriterByPath(
	ctx context.Context,
	path string,
	baseTransport http.RoundTripper,
	dryRun bool,
) (io.WriteCloser, error) {
	if dryRun {
		return &discard{}, nil
	}

	if !isGsutilPath(path) {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, fmt.Errorf("mkdir: %w", err)
		}

		tokenFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return nil, fmt.Errorf("open token file: %w", err)
		}

		return tokenFile, nil
	}

	writer, err := NewWriter(ctx, path, baseTransport)
	if err != nil {
		return nil, fmt.Errorf("new gcs writer: %w", err)
	}
	return writer, nil
}
