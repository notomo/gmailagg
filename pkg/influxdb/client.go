package influxdb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

func NewClient(
	serverURL string,
	authToken string,
	baseTransport http.RoundTripper,
) influxdb2.Client {
	opts := influxdb2.DefaultOptions()
	return influxdb2.NewClientWithOptions(
		serverURL,
		authToken,
		opts.SetHTTPClient(&http.Client{
			Timeout:   20 * time.Second,
			Transport: baseTransport,
		}),
	)
}

type Writer struct {
	api  api.WriteAPI
	errs []error
}

func NewWriter(
	client influxdb2.Client,
	org string,
	bucket string,
	dryRunWriter io.Writer,
) *Writer {
	var api api.WriteAPI
	if dryRunWriter != nil {
		api = NewDryRunWriter(dryRunWriter)
	} else {
		api = client.WriteAPI(org, bucket)
	}

	errs := []error{}
	go func() {
		for err := range api.Errors() {
			errs = append(errs, err)
		}
	}()

	return &Writer{
		api:  api,
		errs: errs,
	}
}

type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]any
	At          time.Time
}

func (w *Writer) Write(ctx context.Context, points ...Point) {
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
	w.api.Flush()
	return errors.Join(w.errs...)
}
