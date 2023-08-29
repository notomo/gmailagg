package extractor

import (
	"log/slog"
	"regexp"
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

func makeMatchMap(
	regex *regexp.Regexp,
	body string,
	maxMatchCount int,
) map[string][]string {
	matchMap := map[string][]string{}
	captureNames := regex.SubexpNames()
	allMatches := regex.FindAllStringSubmatch(body, maxMatchCount)
	for _, matches := range allMatches {
		matchesCount := len(matches)
		for i, name := range captureNames {
			if i == 0 || matchesCount <= i {
				continue
			}
			if _, ok := matchMap[name]; !ok {
				matchMap[name] = []string{}
			}
			matchMap[name] = append(matchMap[name], matches[i])
		}
	}
	return matchMap
}

func resolve(
	mappingName string,
	mappings map[string]RuleMapping,
	matchMap map[string][]string,
	fields map[string]any,
	hiddenFields map[string]any,
	tags map[string]string,
) error {
	mapping, ok := mappings[mappingName]
	if !ok {
		return nil
	}

	matches := matchMap[mappingName]
	for _, rawMatch := range matches {
		match := mapping.Replace(rawMatch)
		switch mapping.Type {
		case RuleMappingTypeField:
			v, err := mapping.FieldValue(match, fields[mappingName])
			if err != nil {
				return err
			}
			fields[mappingName] = v
		case RuleMappingTypeHiddenField:
			v, err := mapping.FieldValue(match, hiddenFields[mappingName])
			if err != nil {
				return err
			}
			hiddenFields[mappingName] = v
		case RuleMappingTypeTag:
			tags[mappingName] = match
		}
	}

	return nil
}

func toExtractor(
	measurementName string,
	query string,
	rules []AggregationRule,
	baseTags map[string]string,
) (*Extractor, error) {
	logger := slog.Default()

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
			logger.Debug("message", "body", body)

			fields := map[string]any{}
			hiddenFields := map[string]any{}
			tags := map[string]string{}
			for k, v := range baseTags {
				tags[k] = v
			}

			matchMap := makeMatchMap(regex, body, rule.MatchMaxCount)
			for mappingName := range rule.Mappings {
				if err := resolve(
					mappingName,
					rule.Mappings,
					matchMap,
					fields,
					hiddenFields,
					tags,
				); err != nil {
					return nil, err
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
