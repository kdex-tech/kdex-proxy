package roles

import (
	"fmt"

	"kdex.dev/proxy/internal/cel"
	"kdex.dev/proxy/internal/config"
)

type RoleEvaluator struct {
	Config    *config.Config
	Evaluator *cel.Evaluator
}

func NewRoleEvaluator(evaluator *cel.Evaluator, config *config.Config) *RoleEvaluator {
	return &RoleEvaluator{Evaluator: evaluator, Config: config}
}

func (e *RoleEvaluator) EvaluateRoles(data map[string]interface{}) ([]string, error) {
	result, err := e.Evaluator.Evaluate(e.Config.Authz.Roles.Expression, data)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	typedResult, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expression must evaluate to a list of strings")
	}

	roles := make([]string, len(typedResult))
	for i, v := range typedResult {
		roles[i] = fmt.Sprint(v)
	}

	return roles, nil
}
