package extractor

import (
	"fmt"
	"regexp"
	"strings"
)

var expressionOperator = regexp.MustCompile("[+-]")

func (m *RuleMapping) ExpressionFields() []string {
	return expressionOperator.Split(m.Expression, -1)
}

func resolveExpression[T int | float64](
	expression string,
	raw string,
	fields map[string]any,
	hiddenValues map[string]any,
	oldValue any,
	parse func(string) (T, error),
) (*T, error) {
	if expression == "" {
		v, err := parse(raw)
		if err != nil {
			return nil, err
		}
		if old, ok := oldValue.(T); ok {
			v += old
		}
		return &v, nil
	}

	expression = expressionOperator.ReplaceAllStringFunc(expression, func(s string) string {
		return fmt.Sprintf(" %s ", s)
	})
	var result T
	var factor T = 1
	for _, atom := range strings.Fields(expression) {
		switch atom {
		case "+":
			factor = 1
		case "-":
			factor = -1
		default:
			var value T
			if v, ok := fields[atom]; ok {
				value = v.(T)
			} else if v, ok := hiddenValues[atom]; ok {
				value = v.(T)
			} else {
				return nil, nil
			}
			result += value * factor
		}
	}

	return &result, nil
}
