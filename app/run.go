package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/notomo/gmailagg/app/extractor"
	"github.com/notomo/gmailagg/pkg/gmailext"
	"github.com/notomo/gmailagg/pkg/influxdb"
)

func Run(
	ctx context.Context,
	credentialsJsonPath string,
	tokenFilePath string,
	measurements []extractor.Measurement,
	influxdbServerURL string,
	influxdbAuthToken string,
	influxdbOrg string,
	influxdbBucket string,
	baseTransport http.RoundTripper,
	dryRunWriter io.Writer,
) (retErr error) {
	influxdbClient := influxdb.NewClient(
		influxdbServerURL,
		influxdbAuthToken,
		baseTransport,
	)
	defer influxdbClient.Close()

	influxdbWriter := influxdb.NewWriter(
		influxdbClient,
		influxdbOrg,
		influxdbBucket,
		dryRunWriter,
	)
	defer func() {
		if err := influxdbWriter.Flush(ctx); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("flush influxdb write: %w", err))
		}
	}()

	service, err := gmailext.NewService(ctx, credentialsJsonPath, tokenFilePath, baseTransport)
	if err != nil {
		return fmt.Errorf("new gmail service: %w", err)
	}

	for _, measurement := range measurements {
		if err := extractor.Do(
			ctx,
			service,
			"me",
			measurement,
			func(ctx context.Context, points ...influxdb.Point) error {
				influxdbWriter.Write(ctx, points...)
				return nil
			},
		); err != nil {
			return fmt.Errorf("extractor do: %w", err)
		}
	}

	return nil
}
