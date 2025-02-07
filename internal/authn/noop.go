package authn

import (
	"net/http"
)

type NoOpAuthValidator struct{}

func (v *NoOpAuthValidator) Register(mux *http.ServeMux) {
	// noop
}

func (v *NoOpAuthValidator) Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler) {
	return nil
}
