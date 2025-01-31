package transform

import "net/http"

type Transformer interface {
	Transform(r *http.Response, body *[]byte) error
	ShouldTransform(r *http.Response) bool
}

type AggregatedTransformer struct {
	Transformer
	Transformers []Transformer
}

func (t *AggregatedTransformer) Transform(r *http.Response, body *[]byte) error {
	for _, transformer := range t.Transformers {
		if err := transformer.Transform(r, body); err != nil {
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
