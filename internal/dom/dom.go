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

package dom

import (
	"bytes"

	"golang.org/x/net/html"
)

func collectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf)
	}
}

func FindElementByName(name string, doc *html.Node, predicate func(n *html.Node) bool) *html.Node {
	var find func(*html.Node) *html.Node
	find = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode && n.Data == name {
			if predicate != nil && !predicate(n) {
				return nil
			}
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := find(c); result != nil {
				return result
			}
		}
		return nil
	}
	return find(doc)
}

func GetNodeText(node *html.Node) []byte {
	var buf bytes.Buffer
	collectText(node, &buf)
	return buf.Bytes()
}
