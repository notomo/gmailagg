package extractor

import (
	"fmt"
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

	Target   TargetType
	Pattern  string
	Mappings map[string]RuleMapping
}

type TargetType string

var (
	TargetTypeBody = TargetType("body")
)

type RuleType string

var (
	RuleTypeRegexp = RuleType("regexp")
)

type RuleMapping struct {
	Type      RuleMappingType          `json:"type"`
	DataType  RuleMappingFieldDataType `json:"dataType"`
	Replacers []Replacer               `json:"replacers"`
}

type Replacer struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func (r *Replacer) Apply(s string) string {
	return strings.ReplaceAll(s, r.Old, r.New)
}

func (m *RuleMapping) Replace(s string) string {
	for _, replacer := range m.Replacers {
		s = replacer.Apply(s)
	}
	return s
}

func (m *RuleMapping) FieldValue(raw string) (any, error) {
	switch m.DataType {
	case RuleMappingFieldDataTypeFloat:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	case RuleMappingFieldDataTypeInteger:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, err
		}
		return v, nil
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

type RuleMappingType string

var (
	RuleMappingTypeField = RuleMappingType("field")
	RuleMappingTypeTag   = RuleMappingType("tag")
)

type RuleMappingFieldDataType string

var (
	RuleMappingFieldDataTypeFloat   = RuleMappingFieldDataType("float")
	RuleMappingFieldDataTypeInteger = RuleMappingFieldDataType("integer")
	RuleMappingFieldDataTypeString  = RuleMappingFieldDataType("string")
	RuleMappingFieldDataTypeBoolean = RuleMappingFieldDataType("boolean")
)
