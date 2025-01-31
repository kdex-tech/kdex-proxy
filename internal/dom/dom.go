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
