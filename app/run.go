package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/notomo/gmailagg/app/extractor"
	"github.com/notomo/gmailagg/pkg/gcsext"
	"github.com/notomo/gmailagg/pkg/gmailext"
	"github.com/notomo/gmailagg/pkg/influxdb"
	"google.golang.org/api/gmail/v1"
)

func Run(
	ctx context.Context,
	gmailCredentials string,
	tokenFilePath string,
	measurements []extractor.Measurement,
	influxdbServerURL string,
	influxdbAuthToken string,
	influxdbOrg string,
	influxdbBucket string,
	baseTransport http.RoundTripper,
	dryRunWriter io.Writer,
) (retErr error) {
	influxdbWriter := influxdb.NewWriter(
		influxdbServerURL,
		influxdbAuthToken,
		influxdbOrg,
		influxdbBucket,
		dryRunWriter,
		baseTransport,
	)
	defer func() {
		if err := influxdbWriter.Flush(ctx); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("flush influxdb write: %w", err))
		}
	}()

	tokenReader, err := gcsext.NewReaderByPath(ctx, tokenFilePath, baseTransport)
	if err != nil {
		return fmt.Errorf("new token reader: %w", err)
	}
	defer func() {
		if err := tokenReader.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close token reader: %w", err))
		}
	}()

	service, err := gmailext.NewService(ctx, gmailCredentials, tokenReader, baseTransport)
	if err != nil {
		return fmt.Errorf("new gmail service: %w", err)
	}

	extractors, err := extractor.List(measurements)
	if err != nil {
		return fmt.Errorf("extractor list: %w", err)
	}

	logger := slog.Default()
	var allCount int64
	for _, e := range extractors {
		if err := gmailext.Iter(
			ctx,
			service,
			"me",
			e.Query,
			func(ctx context.Context, message *gmail.Message) (bool, error) {
				points, err := e.Convert(message)
				if err != nil {
					return false, err
				}

				count := len(points)
				if count > 0 {
					logger.Info("writing points", "count", count)
				}
				atomic.AddInt64(&allCount, int64(count))

				influxdbWriter.Write(ctx, points...)
				return true, nil
			},
		); err != nil {
			return fmt.Errorf("gmailext iter: %w", err)
		}
	}
	logger.Info("writing points", "allCount", allCount)

	return nil
}
