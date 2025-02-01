// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package importmap

import (
	"bytes"
	"reflect"
	"testing"

	"golang.org/x/net/html"
)

func TestImportMapManager_Mutate(t *testing.T) {
	tests := []struct {
		name    string
		docNode *html.Node
		imports map[string]string
	}{
		{
			name:    "no doc node",
			docNode: toDoc(""),
			imports: map[string]string{
				"@kdex-ui": "/~/m/kdex-ui/index.js",
			},
		},
		{
			name:    "no script node",
			docNode: toDoc("<html></html>"),
			imports: map[string]string{
				"@kdex-ui": "/~/m/kdex-ui/index.js",
			},
		},
		{
			name:    "mutate importmap",
			docNode: toDoc(`<html><script type="importmap"></script></html>`),
			imports: map[string]string{
				"@kdex-ui": "/~/m/kdex-ui/index.js",
			},
		},
		{
			name:    "mutate importmap with existing imports",
			docNode: toDoc(`<html><script type="importmap">{"imports":{"@foo/bar":"/foo/bar.js"}}</script></html>`),
			imports: map[string]string{
				"@foo/bar": "/foo/bar.js",
				"@kdex-ui": "/~/m/kdex-ui/index.js",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importMapInstance := Instance(tt.docNode).WithMutator(
				func(importMap *ImportMap) {
					importMap.Imports["@kdex-ui"] = "/~/m/kdex-ui/index.js"
				},
			)
			importMapInstance.Mutate()
			if tt.imports != nil {
				if !reflect.DeepEqual(importMapInstance.importMap.Imports, tt.imports) {
					t.Errorf("ImportMapManager.Mutate() = %v, want %v", importMapInstance.importMap.Imports, tt.imports)
				}
			}
		})
	}
}

func toDoc(body string) *html.Node {
	doc, _ := html.Parse(bytes.NewReader([]byte(body)))
	return doc
}
