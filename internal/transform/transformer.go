package transform

import "net/http"

type Transformer interface {
	Transform(body *[]byte) error
	ShouldTransform(r *http.Response) bool
}
