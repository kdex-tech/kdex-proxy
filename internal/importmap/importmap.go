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
	"encoding/json"

	"golang.org/x/net/html"
)

type Mutator func(importMap *ImportMap)

type ImportMapManager struct {
	docNode   *html.Node
	mapNode   *html.Node
	importMap ImportMap
	mutator   Mutator
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

func (importMapManager *ImportMapManager) WithMutator(mutator func(importMap *ImportMap)) *ImportMapManager {
	importMapManager.mutator = mutator
	return importMapManager
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
	if importMapManager.importMap.Integrity == nil {
		importMapManager.importMap.Integrity = make(map[string]string)
	}
	if importMapManager.importMap.Scopes == nil {
		importMapManager.importMap.Scopes = make(map[string]map[string]string)
	}

	if importMapManager.mutator != nil {
		importMapManager.mutator(&importMapManager.importMap)
	}

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
