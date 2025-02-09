package transform

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Transformer interface {
	Transform(r *http.Response, doc *html.Node) error
	ShouldTransform(r *http.Response) bool
}

type AggregatedTransformer struct {
	Transformer
	Transformers []Transformer
}

func (t *AggregatedTransformer) Transform(r *http.Response, doc *html.Node) error {
	for _, transformer := range t.Transformers {
		if !transformer.ShouldTransform(r) {
			continue
		}
		if err := transformer.Transform(r, doc); err != nil {
			return err
		}
	}
	return nil
}

func (t *AggregatedTransformer) ShouldTransform(r *http.Response) bool {
	return HtmlTransformCheck(r)
}

func HtmlTransformCheck(r *http.Response) bool {
	// Check if response is HTML and not streaming
	contentType := r.Header.Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isStreaming := r.Header.Get("Transfer-Encoding") == "chunked"

	if !isHTML || isStreaming {
		return false
	}

	return true
}
