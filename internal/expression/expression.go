// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
