package influxdb

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Writer struct {
	client influxdb2.Client
	api    api.WriteAPI
	errCh  <-chan error
}

func NewWriter(
	serverURL string,
	authToken string,
	org string,
	bucket string,
	dryRunWriter io.Writer,
	baseTransport http.RoundTripper,
) *Writer {
	opts := influxdb2.DefaultOptions()
	client := influxdb2.NewClientWithOptions(
		serverURL,
		authToken,
		opts.SetHTTPClient(&http.Client{
			Timeout:   20 * time.Second,
			Transport: baseTransport,
		}),
	)

	var api api.WriteAPI
	if dryRunWriter != nil {
		api = NewDryRunWriter(dryRunWriter)
	} else {
		api = client.WriteAPI(org, bucket)
	}

	return &Writer{
		client: client,
		api:    api,
		errCh:  api.Errors(),
	}
}

type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]any
	At          time.Time
}

func (w *Writer) Write(ctx context.Context, points ...Point) {
	count := len(points)
	if count > 0 {
		log.Printf("writing points: count=%d", count)
	}
	for _, p := range points {
		w.api.WritePoint(influxdb2.NewPoint(
			p.Measurement,
			p.Tags,
			p.Fields,
			p.At,
		))
	}
}

func (w *Writer) Flush(ctx context.Context) error {
	errs := []error{}

	wait := make(chan struct{})
	go func() {
		for err := range w.errCh {
			errs = append(errs, err)
		}
		wait <- struct{}{}
	}()

	w.api.Flush()
	w.client.Close()
	<-wait

	return errors.Join(errs...)
}
