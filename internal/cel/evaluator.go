package cel

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type Evaluator struct {
	env *cel.Env
}

func NewEvaluator() (*Evaluator, error) {
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("this", decls.NewMapType(decls.String, decls.Any)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %v", err)
	}

	return &Evaluator{env: env}, nil
}

func (e *Evaluator) Evaluate(expression string, data map[string]interface{}) (any, error) {
	ast, iss := e.env.Compile(expression)
	if iss.Err() != nil {
		return nil, fmt.Errorf("failed to compile expression: %v", iss.Err())
	}

	prg, err := e.env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to create program: %v", err)
	}

	out, _, err := prg.Eval(map[string]interface{}{
		"this": data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	return out.Value(), nil
}
