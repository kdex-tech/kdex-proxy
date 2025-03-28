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

package transform

import (
	"net/http"

	"golang.org/x/net/html"
)

type Transformer interface {
	Transform(r *http.Response, doc *html.Node) error
}

type AggregatedTransformer struct {
	Transformer
	Transformers []Transformer
}

func (t *AggregatedTransformer) Transform(r *http.Response, doc *html.Node) error {
	for _, transformer := range t.Transformers {
		if err := transformer.Transform(r, doc); err != nil {
			return err
		}
	}
	return nil
}
