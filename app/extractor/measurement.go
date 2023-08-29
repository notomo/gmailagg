package extractor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Measurement struct {
	Name         string
	Aggregations []Aggregation
}

type Aggregation struct {
	Query string
	Rules []AggregationRule
	Tags  map[string]string
}

type AggregationRule struct {
	Type RuleType

	Target        TargetType
	Pattern       string
	MatchMaxCount int
	Mappings      map[string]RuleMapping
}

type TargetType string

var (
	TargetTypeBody = TargetType("body")
)

type RuleType string

var (
	RuleTypeRegexp = RuleType("regexp")
)

type Replacer struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func (r *Replacer) Apply(s string) string {
	return strings.ReplaceAll(s, r.Old, r.New)
}

type RuleMapping struct {
	Type       RuleMappingType          `json:"type"`
	DataType   RuleMappingFieldDataType `json:"dataType"`
	Replacers  []Replacer               `json:"replacers"`
	Expression string                   `json:"expression"`
}

func (m *RuleMapping) ExpressionFields() []string {
	return expressionOperator.Split(m.Expression, -1)
}

func (m *RuleMapping) Replace(s string) string {
	for _, replacer := range m.Replacers {
		s = replacer.Apply(s)
	}
	return s
}

func (m *RuleMapping) FieldValue(
	raw string,
	oldValue any,
	fields map[string]any,
	hiddenValues map[string]any,
) (any, error) {
	switch m.DataType {
	case RuleMappingFieldDataTypeFloat:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, err
		}
		if old, ok := oldValue.(float64); ok {
			v += old
		}
		return v, nil
	case RuleMappingFieldDataTypeInteger:
		v, resolved, err := resolveIntegerExpression(m.Expression, raw, fields, hiddenValues)
		if err != nil {
			return nil, err
		}
		if old, ok := oldValue.(int); ok {
			v += old
		}
		if resolved {
			return v, nil
		}
		return nil, nil
	case RuleMappingFieldDataTypeString:
		return raw, nil
	case RuleMappingFieldDataTypeBoolean:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, fmt.Errorf("unexpected data_type: %s", m.DataType)
}

var expressionOperator = regexp.MustCompile("[+-]")

func resolveIntegerExpression(
	expression string,
	raw string,
	fields map[string]any,
	hiddenValues map[string]any,
) (int, bool, error) {
	if expression == "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return 0, false, err
		}
		return v, true, nil
	}

	expression = expressionOperator.ReplaceAllStringFunc(expression, func(s string) string {
		return fmt.Sprintf(" %s ", s)
	})
	result := 0
	factor := 1
	for _, atom := range strings.Fields(expression) {
		switch atom {
		case "+":
			factor = 1
		case "-":
			factor = -1
		default:
			var value int
			if v, ok := fields[atom]; ok {
				value = v.(int)
			} else if v, ok := hiddenValues[atom]; ok {
				value = v.(int)
			} else {
				return 0, false, nil
			}
			result += value * factor
		}
	}

	return result, true, nil
}

type RuleMappingType string

var (
	RuleMappingTypeField       = RuleMappingType("field")
	RuleMappingTypeTag         = RuleMappingType("tag")
	RuleMappingTypeHiddenValue = RuleMappingType("hidden")
)

type RuleMappingFieldDataType string

var (
	RuleMappingFieldDataTypeFloat   = RuleMappingFieldDataType("float")
	RuleMappingFieldDataTypeInteger = RuleMappingFieldDataType("integer")
	RuleMappingFieldDataTypeString  = RuleMappingFieldDataType("string")
	RuleMappingFieldDataTypeBoolean = RuleMappingFieldDataType("boolean")
)
