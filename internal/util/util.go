package util

import "net/http"

func GetScheme(r *http.Request) string {
	if r.URL.Scheme != "" {
		return r.URL.Scheme
	}
	return "http"
}
