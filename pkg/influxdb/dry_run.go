package influxdb

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	protocol "github.com/influxdata/line-protocol"
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
}

type dryRunPoint struct {
	Measurement string
	Tags        []*protocol.Tag
	Fields      []*protocol.Field
	At          time.Time
}

func (w *DryRunWriter) WritePoint(point *write.Point) {
	var b bytes.Buffer
	encorder := json.NewEncoder(&b)
	encorder.SetIndent("", "  ")
	if err := encorder.Encode(dryRunPoint{
		Measurement: point.Name(),
		Tags:        point.TagList(),
		Fields:      point.FieldList(),
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
