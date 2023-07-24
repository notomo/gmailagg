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
	"google.golang.org/api/gmail/v1"
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

	extractors, err := extractor.List(measurements)
	if err != nil {
		return fmt.Errorf("extractor list: %w", err)
	}

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
				influxdbWriter.Write(ctx, points...)
				return true, nil
			},
		); err != nil {
			return fmt.Errorf("gmailext iter: %w", err)
		}
	}

	return nil
}
