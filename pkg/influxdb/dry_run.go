package influxdb

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type DryRunWriter struct {
	errCh        chan error
	outputWriter io.Writer
	points       [][]byte
}

func NewDryRunWriter(outputWriter io.Writer) *DryRunWriter {
	return &DryRunWriter{
		errCh:        make(chan error, 1),
		outputWriter: outputWriter,
		points:       [][]byte{},
	}
}

func (w *DryRunWriter) Errors() <-chan error {
	return w.errCh
}

func (w *DryRunWriter) Flush() {
	for _, point := range w.points {
		if _, err := w.outputWriter.Write(point); err != nil {
			w.errCh <- err
		}
	}
	close(w.errCh)
}

type dryRunTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type dryRunField struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type dryRunPoint struct {
	Measurement string        `json:"measurement"`
	Tags        []dryRunTag   `json:"tags"`
	Fields      []dryRunField `json:"fields"`
	At          time.Time     `json:"at"`
}

func (w *DryRunWriter) WritePoint(point *write.Point) {
	tags := []dryRunTag{}
	for _, tag := range point.TagList() {
		tags = append(tags, dryRunTag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	fields := []dryRunField{}
	for _, tag := range point.FieldList() {
		fields = append(fields, dryRunField{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	var b bytes.Buffer
	encorder := json.NewEncoder(&b)
	encorder.SetIndent("", "  ")
	if err := encorder.Encode(dryRunPoint{
		Measurement: point.Name(),
		Tags:        tags,
		Fields:      fields,
		At:          point.Time(),
	}); err != nil {
		w.errCh <- err
	}
	w.points = append(w.points, b.Bytes())
}

func (w *DryRunWriter) SetWriteFailedCallback(cb api.WriteFailedCallback) {
	panic("unimplemented")
}

func (w *DryRunWriter) WriteRecord(line string) {
	panic("unimplemented")
}

var _ api.WriteAPI = &DryRunWriter{}
