package expression

import (
	"fmt"
	"log"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type Evaluator struct {
	env *cel.Env
}

func NewEvaluator() *Evaluator {
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("data", decls.NewMapType(decls.String, decls.Any)),
		),
	)
	if err != nil {
		log.Fatalf("failed to create CEL environment: %v", err)
	}

	return &Evaluator{env: env}
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
		"data": data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	return out.Value(), nil
}
