package mux

import "net/http"

type Muxable interface {
	Register(mux *http.ServeMux)
}
