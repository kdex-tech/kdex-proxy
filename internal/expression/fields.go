package expression

import (
	"fmt"

	"kdex.dev/proxy/internal/config"
)

type FieldEvaluator struct {
	Config    *config.Config
	Evaluator *Evaluator
}

func NewFieldEvaluator(config *config.Config) *FieldEvaluator {
	evaluator := NewEvaluator()
	return &FieldEvaluator{Evaluator: evaluator, Config: config}
}

func (e *FieldEvaluator) EvaluatePrincipal(data map[string]interface{}) (string, error) {
	result, err := e.Evaluator.Evaluate(e.Config.Expressions.Principal, data)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate expression: %v", err)
	}

	switch v := result.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("expression must evaluate to a string; got %T for %v", v, result)
	}
}

func (e *FieldEvaluator) EvaluateRoles(data map[string]interface{}) ([]string, error) {
	result, err := e.Evaluator.Evaluate(e.Config.Expressions.Roles, data)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	switch v := result.(type) {
	case []interface{}:
		roles := make([]string, len(v))
		for i, v := range v {
			roles[i] = fmt.Sprint(v)
		}
		return roles, nil
	case []string:
		return v, nil
	default:
		return nil, fmt.Errorf("expression must evaluate to a list of strings; got %T for %v", v, result)
	}
}
