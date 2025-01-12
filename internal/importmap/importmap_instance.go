package importmap

import (
	"bytes"
	"encoding/json"
	"log"

	"golang.org/x/net/html"
)

type ImportMapInstance struct {
	docNode    *html.Node
	mapNode    *html.Node
	importMap  ImportMap
	mutator    ImportMapMutator
	moduleBody string
}

func Parse(body *[]byte) (*ImportMapInstance, error) {
	doc, err := html.Parse(bytes.NewReader(*body))
	if err != nil {
		return nil, err
	}

	importMapInstance := Instance(doc)

	return importMapInstance, nil
}

func (importMapInstance *ImportMapInstance) Return(body *[]byte) error {
	var buf bytes.Buffer
	if err := html.Render(&buf, importMapInstance.docNode); err != nil {
		log.Printf("Error rendering modified HTML: %v", err)
		return err
	}
	*body = buf.Bytes()
	return nil
}

func Instance(doc *html.Node) *ImportMapInstance {
	var importMapInstance ImportMapInstance
	importMapInstance.docNode = doc
	importMapInstance.mapNode = findElementByName("script", doc, func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == "type" && a.Val == "importmap" {
				return true
			}
		}

		return false
	})
	return &importMapInstance
}

func (importMapInstance *ImportMapInstance) WithMutator(mutator func(importMap *ImportMap)) *ImportMapInstance {
	importMapInstance.mutator = mutator
	return importMapInstance
}

func (importMapInstance *ImportMapInstance) WithModuleBody(moduleBody string) *ImportMapInstance {
	importMapInstance.moduleBody = moduleBody
	return importMapInstance
}

func (importMapInstance *ImportMapInstance) Mutate() bool {
	if importMapInstance.mapNode == nil {
		headNode := findElementByName("head", importMapInstance.docNode, nil)
		if headNode == nil {
			return false
		}

		importMapInstance.mapNode = &html.Node{
			Type: html.ElementNode,
			Data: "script",
			Attr: []html.Attribute{
				{Key: "type", Val: "importmap"},
			},
		}

		headNode.InsertBefore(importMapInstance.mapNode, headNode.FirstChild)
	}

	bytes := getNodeText(importMapInstance.mapNode)

	json.Unmarshal(bytes, &importMapInstance.importMap)

	if importMapInstance.importMap.Imports == nil {
		importMapInstance.importMap.Imports = make(map[string]string)
	}
	if importMapInstance.importMap.Integrity == nil {
		importMapInstance.importMap.Integrity = make(map[string]string)
	}
	if importMapInstance.importMap.Scopes == nil {
		importMapInstance.importMap.Scopes = make(map[string]map[string]string)
	}

	if importMapInstance.mutator != nil {
		importMapInstance.mutator(&importMapInstance.importMap)
	}

	mapBytes, _ := json.Marshal(importMapInstance.importMap)

	newTextNode := &html.Node{
		Type: html.TextNode,
		Data: string(mapBytes),
	}

	// Remove existing text nodes
	for c := importMapInstance.mapNode.FirstChild; c != nil; {
		next := c.NextSibling
		importMapInstance.mapNode.RemoveChild(c)
		c = next
	}

	// Add new text node
	importMapInstance.mapNode.AppendChild(newTextNode)

	// Append a new import script node to the bottom of the body
	bodyNode := findElementByName("body", importMapInstance.docNode, nil)
	if bodyNode != nil {
		scriptNode := &html.Node{
			Type: html.ElementNode,
			Data: "script",
			Attr: []html.Attribute{
				{Key: "type", Val: "module"},
			},
		}

		scriptNode.AppendChild(&html.Node{
			Type: html.TextNode,
			Data: importMapInstance.moduleBody,
		})

		bodyNode.AppendChild(scriptNode)
	}

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

func findElementByName(name string, doc *html.Node, predicate func(n *html.Node) bool) *html.Node {
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

func getNodeText(node *html.Node) []byte {
	var buf bytes.Buffer
	collectText(node, &buf)
	return buf.Bytes()
}
