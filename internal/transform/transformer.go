package transform

import (
	"net/http"

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
		if err := transformer.Transform(r, doc); err != nil {
			return err
		}
	}
	return nil
}

func (t *AggregatedTransformer) ShouldTransform(r *http.Response) bool {
	for _, transformer := range t.Transformers {
		if transformer.ShouldTransform(r) {
			return true
		}
	}
	return false
}
