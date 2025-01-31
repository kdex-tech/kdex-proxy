package dom

import (
	"bytes"
	"reflect"
	"testing"

	"golang.org/x/net/html"
)

func Test_collectText(t *testing.T) {
	type args struct {
		n   *html.Node
		buf *bytes.Buffer
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectText(tt.args.n, tt.args.buf)
		})
	}
}

func TestFindElementByName(t *testing.T) {
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
			if got := FindElementByName("script", tt.args.doc, func(n *html.Node) bool {
				return n.Attr[0].Val == "importmap"
			}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findImportMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNodeText(t *testing.T) {
	type args struct {
		node *html.Node
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNodeText(tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeText() = %v, want %v", got, tt.want)
			}
		})
	}
}
