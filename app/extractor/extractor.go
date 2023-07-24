package extractor

import (
	"regexp"
	"strconv"
	"time"

	"github.com/notomo/gmailagg/pkg/gmailext"
	"github.com/notomo/gmailagg/pkg/influxdb"
	"google.golang.org/api/gmail/v1"
)

type Extractor struct {
	Query   string
	Convert func(*gmail.Message) ([]influxdb.Point, error)
}

func List(
	measurements []Measurement,
) ([]Extractor, error) {
	extractors := []Extractor{}
	for _, measurement := range measurements {
		for _, aggregation := range measurement.Aggregations {
			e, err := toExtractor(measurement.Name, aggregation.Query, aggregation.Rules, aggregation.Tags)
			if err != nil {
				return nil, err
			}
			extractors = append(extractors, *e)
		}
	}
	return extractors, nil
}

func toExtractor(
	measurementName string,
	query string,
	rules []AggregationRule,
	baseTags map[string]string,
) (*Extractor, error) {
	funcs := []func(*gmail.Message) (*influxdb.Point, error){}
	for _, rule := range rules {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, err
		}

		f := func(message *gmail.Message) (*influxdb.Point, error) {
			body, err := gmailext.StringBody(message)
			if err != nil {
				return nil, err
			}

			fields := map[string]any{}
			tags := map[string]string{}
			for k, v := range baseTags {
				tags[k] = v
			}

			matches := regex.FindStringSubmatch(body)
			matchesCount := len(matches)
			for i, name := range regex.SubexpNames() {
				if i == 0 || name == "" || matchesCount <= i {
					continue
				}

				match := matches[i]
				mapping, ok := rule.Mappings[name]
				if !ok {
					continue
				}

				switch mapping.Type {
				case RuleMappingTypeField:
					v, err := strconv.Atoi(match)
					if err != nil {
						return nil, err
					}
					fields[name] = v
				case RuleMappingTypeTag:
					tags[name] = match
				}
			}

			if len(fields) == 0 {
				return nil, nil
			}

			return &influxdb.Point{
				Measurement: measurementName,
				Tags:        tags,
				Fields:      fields,
				At:          time.UnixMilli(message.InternalDate),
			}, nil
		}
		funcs = append(funcs, f)
	}

	convert := func(message *gmail.Message) ([]influxdb.Point, error) {
		points := []influxdb.Point{}
		for _, f := range funcs {
			point, err := f(message)
			if err != nil {
				return nil, err
			}
			if point != nil {
				points = append(points, *point)
			}
		}
		return points, nil
	}
	return &Extractor{
		Query:   query,
		Convert: convert,
	}, nil
}
