package extractor

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
	Type RuleMappingType
}

type RuleMappingType string

var (
	RuleMappingTypeField = RuleMappingType("field")
	RuleMappingTypeTag   = RuleMappingType("tag")
)
