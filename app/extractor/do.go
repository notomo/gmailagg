package extractor

import (
	"context"

	"github.com/notomo/gmailagg/pkg/gmailext"
	"github.com/notomo/gmailagg/pkg/influxdb"
	"google.golang.org/api/gmail/v1"
)

func Do(
	ctx context.Context,
	service *gmail.Service,
	userID string,
	measurement Measurement,
	each func(context.Context, ...influxdb.Point) error,
) error {
	for _, aggregation := range measurement.Aggregations {
		convert, err := toExtractor(measurement.Name, aggregation.Rules, aggregation.Tags)
		if err != nil {
			return err
		}

		if err := gmailext.Iter(
			ctx,
			service,
			userID,
			aggregation.Query,
			func(ctx context.Context, message *gmail.Message) (bool, error) {
				points, err := convert(message)
				if err != nil {
					return false, err
				}
				if err := each(ctx, points...); err != nil {
					return false, err
				}
				return true, nil
			},
		); err != nil {
			return err
		}
	}
	return nil
}
