package extractor

import (
	"time"
)

type Point struct {
	Measurement string            `json:"measurement"`
	Tags        map[string]string `json:"tags"`
	Fields      map[string]any    `json:"fields"`
	At          time.Time         `json:"at"`
}
