package importmap

import (
	"bytes"
	"encoding/json"

	"golang.org/x/net/html"
)

type ImportMapManager struct {
	docNode   *html.Node
	mapNode   *html.Node
	importMap ImportMap
}

type ImportMap struct {
	Imports   map[string]string            `json:"imports,omitempty"`
	Integrity map[string]string            `json:"integrity,omitempty"`
	Scopes    map[string]map[string]string `json:"scopes,omitempty"`
}

func Manager(doc *html.Node) *ImportMapManager {
	var importMapManager ImportMapManager
	importMapManager.docNode = doc
	importMapManager.mapNode = findImportMap(doc)
	return &importMapManager
}

func (importMapManager *ImportMapManager) Mutate() bool {
	if importMapManager.mapNode == nil {
		return false
	}

	bytes := getNodeText(importMapManager.mapNode)

	json.Unmarshal(bytes, &importMapManager.importMap)

	if importMapManager.importMap.Imports == nil {
		importMapManager.importMap.Imports = make(map[string]string)
	}

	importMapManager.importMap.Imports["@kdex-ui"] = "/_/kdex-ui.js"

	mapBytes, _ := json.Marshal(importMapManager.importMap)

	newTextNode := &html.Node{
		Type: html.TextNode,
		Data: string(mapBytes),
	}

	// Remove existing text nodes
	for c := importMapManager.mapNode.FirstChild; c != nil; {
		next := c.NextSibling
		importMapManager.mapNode.RemoveChild(c)
		c = next
	}

	// Add new text node
	importMapManager.mapNode.AppendChild(newTextNode)

	return true
}

func collectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf)
	}
}

func findImportMap(doc *html.Node) *html.Node {
	var find func(*html.Node) *html.Node
	find = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, a := range n.Attr {
				if a.Key == "type" && a.Val == "importmap" {
					return n
				}
			}
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

func getNodeText(node *html.Node) []byte {
	var buf bytes.Buffer
	collectText(node, &buf)
	return buf.Bytes()
}
