package extractor

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/notomo/gmailagg/pkg/gmailext"
	"google.golang.org/api/gmail/v1"
)

type Extractor struct {
	Query   string
	Convert func(*gmail.Message) ([]Point, error)
}

func List(
	measurements []Measurement,
) ([]Extractor, error) {
	extractors := []Extractor{}
	for _, measurement := range measurements {
		for _, aggregation := range measurement.Aggregations {
			query := strings.Join([]string{measurement.Query, aggregation.Query}, " ")
			e, err := toExtractor(measurement.Name, query, aggregation.Rules, aggregation.Tags)
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
	mappings map[string]RuleMapping,
	maxMatchCount int,
) (map[string][]string, error) {
	matchMap := map[string][]string{}

	allMatches := regex.FindAllStringSubmatch(body, maxMatchCount)
	if len(allMatches) == 0 {
		return nil, fmt.Errorf("does not matched with body:\n%s", body)
	}

	captureNames := regex.SubexpNames()
	for _, matches := range allMatches {
		matchesCount := len(matches)
		for i, name := range captureNames {
			if i == 0 || matchesCount <= i {
				continue
			}
			if _, ok := matchMap[name]; !ok {
				matchMap[name] = []string{}
			}
			match := matches[i]
			if mapping, ok := mappings[name]; ok {
				match = mapping.Replace(match)
			}
			matchMap[name] = append(matchMap[name], match)
		}
	}

	for name, mapping := range mappings {
		if _, ok := matchMap[name]; !ok && mapping.Expression != "" {
			matchMap[name] = []string{""}
		}
	}

	return matchMap, nil
}

func resolve(
	mappingName string,
	mappings map[string]RuleMapping,
	matchMap map[string][]string,
	fields map[string]any,
	hiddenValues map[string]any,
	tags map[string]string,
) error {
	if _, ok := fields[mappingName]; ok {
		return nil
	}
	if _, ok := hiddenValues[mappingName]; ok {
		return nil
	}
	if _, ok := tags[mappingName]; ok {
		return nil
	}

	mapping, ok := mappings[mappingName]
	if !ok {
		return nil
	}

	for _, fieldName := range mapping.ExpressionFields() {
		if err := resolve(
			fieldName,
			mappings,
			matchMap,
			fields,
			hiddenValues,
			tags,
		); err != nil {
			return err
		}
	}

	matches := matchMap[mappingName]
	for _, match := range matches {
		switch mapping.Type {
		case RuleMappingTypeField:
			v, err := mapping.FieldValue(
				match,
				fields[mappingName],
				fields,
				hiddenValues,
			)
			if err != nil {
				return err
			}
			if v != nil {
				fields[mappingName] = v
			}
		case RuleMappingTypeTag:
			tags[mappingName] = match
		case RuleMappingTypeHiddenValue:
			v, err := mapping.FieldValue(
				match,
				hiddenValues[mappingName],
				fields,
				hiddenValues,
			)
			if err != nil {
				return err
			}
			if v != nil {
				hiddenValues[mappingName] = v
			}
		default:
			return fmt.Errorf("unexpected rule mapping type: %s", mapping.Type)
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

	newlineReplacer := strings.NewReplacer(
		"\r\n", "\n",
		"\r", "\n",
	)

	funcs := []func(*gmail.Message) (*Point, error){}
	for _, rule := range rules {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, err
		}

		f := func(message *gmail.Message) (*Point, error) {
			body, err := gmailext.StringBody(message)
			if err != nil {
				return nil, err
			}
			logger.Debug("message", "body", body)
			body = newlineReplacer.Replace(body)
			body = rule.Replacers.Apply(body)

			fields := map[string]any{}
			hiddenFields := map[string]any{}
			tags := map[string]string{}
			for k, v := range baseTags {
				tags[k] = v
			}

			matchMap, err := makeMatchMap(regex, body, rule.Mappings, rule.MatchMaxCount)
			if err != nil {
				return nil, fmt.Errorf("on pattern %s: %w", rule.Pattern, err)
			}

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

			return &Point{
				Measurement: measurementName,
				Tags:        tags,
				Fields:      fields,
				At:          time.UnixMilli(message.InternalDate),
			}, nil
		}
		funcs = append(funcs, f)
	}

	convert := func(message *gmail.Message) ([]Point, error) {
		points := []Point{}
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
