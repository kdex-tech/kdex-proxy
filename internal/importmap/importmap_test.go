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
		want    bool
	}{
		{
			name:    "no doc node",
			docNode: toDoc(""),
			imports: map[string]string{
				"@kdex-ui": "/_/kdex-ui.js",
			},
			want: true,
		},
		{
			name:    "no script node",
			docNode: toDoc("<html></html>"),
			imports: map[string]string{
				"@kdex-ui": "/_/kdex-ui.js",
			},
			want: true,
		},
		{
			name:    "mutate importmap",
			docNode: toDoc(`<html><script type="importmap"></script></html>`),
			imports: map[string]string{
				"@kdex-ui": "/_/kdex-ui.js",
			},
			want: true,
		},
		{
			name:    "mutate importmap with existing imports",
			docNode: toDoc(`<html><script type="importmap">{"imports":{"@foo/bar":"/foo/bar.js"}}</script></html>`),
			imports: map[string]string{
				"@foo/bar": "/foo/bar.js",
				"@kdex-ui": "/_/kdex-ui.js",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importMapManager := Manager(tt.docNode).WithMutator(
				func(importMap *ImportMap) {
					importMap.Imports["@kdex-ui"] = "/_/kdex-ui.js"
				},
			)
			got := importMapManager.Mutate()
			if got != tt.want {
				t.Errorf("ImportMapManager.Mutate() = %v, want %v", got, tt.want)
			}
			if tt.imports != nil {
				if !reflect.DeepEqual(importMapManager.importMap.Imports, tt.imports) {
					t.Errorf("ImportMapManager.Mutate() = %v, want %v", importMapManager.importMap.Imports, tt.imports)
				}
			}
		})
	}
}

func Test_findImportMap(t *testing.T) {
	type args struct {
		doc *html.Node
	}
	tests := []struct {
		name string
		args args
		want *html.Node
	}{
		{
			name: "script with importmap",
			args: args{
				doc: &html.Node{
					Type: html.ElementNode,
					Data: "script",
					Attr: []html.Attribute{
						{Key: "type", Val: "importmap"},
					},
					FirstChild: &html.Node{
						Type: html.TextNode,
						Data: `{"imports":{"@foo/bar":"/foo/bar.js"}}`,
					},
				},
			},
			want: &html.Node{
				Type: html.ElementNode,
				Data: "script",
				Attr: []html.Attribute{
					{Key: "type", Val: "importmap"},
				},
				FirstChild: &html.Node{
					Type: html.TextNode,
					Data: `{"imports":{"@foo/bar":"/foo/bar.js"}}`,
				},
			},
		},
		{
			name: "no script with importmap",
			args: args{
				doc: &html.Node{
					Type: html.ElementNode,
					Data: "script",
					Attr: []html.Attribute{
						{Key: "type", Val: "text/javascript"},
					},
					FirstChild: &html.Node{
						Type: html.TextNode,
						Data: `{"imports":{"@foo/bar":"/foo/bar.js"}}`,
					},
				},
			},
			want: nil,
		},
		{
			name: "no node",
			args: args{
				doc: &html.Node{
					Type: html.ElementNode,
					Data: "foo",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findImportMap(tt.args.doc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findImportMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toDoc(body string) *html.Node {
	doc, _ := html.Parse(bytes.NewReader([]byte(body)))
	return doc
}
