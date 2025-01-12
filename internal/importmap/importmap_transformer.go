package importmap

import (
	"net/http"
	"strings"
)

type ImportMapTransformer struct {
}

func NewImportMapTransformer() *ImportMapTransformer {
	return &ImportMapTransformer{}
}

func (t *ImportMapTransformer) ShouldTransform(r *http.Response) bool {
	// Check if response is HTML and not streaming
	contentType := r.Header.Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isStreaming := r.Header.Get("Transfer-Encoding") == "chunked"

	if !isHTML || isStreaming {
		return false
	}

	return true
}

func (t *ImportMapTransformer) Transform(body *[]byte) error {
	importMapInstance, err := Parse(body)
	if err != nil {
		return err
	}

	importMapInstance.WithMutator(
		func(importMap *ImportMap) {
			importMap.Imports["@kdex-ui"] = "/~/m/kdex-ui/index.js"
		},
	)

	if !importMapInstance.Mutate() {
		return nil
	}

	if err := importMapInstance.Return(body); err != nil {
		return err
	}

	return nil
}
