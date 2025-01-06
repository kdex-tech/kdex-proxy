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
			want:    false,
		},
		{
			name:    "no script node",
			docNode: toDoc("<html></html>"),
			want:    false,
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
