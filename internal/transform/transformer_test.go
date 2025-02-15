package transform

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHtmlTransformCheck(t *testing.T) {
	tests := []struct {
		name string
		r    *http.Response
		want bool
	}{
		{
			name: "html response",
			r: &http.Response{
				Header: http.Header{"Content-Type": {"text/html"}},
			},
			want: true,
		},
		{
			name: "streaming response",
			r: &http.Response{
				Header: http.Header{"Content-Type": {"text/html"}, "Transfer-Encoding": {"chunked"}},
			},
			want: false,
		},
		{
			name: "non-html response",
			r: &http.Response{
				Header: http.Header{"Content-Type": {"application/json"}},
			},
			want: false,
		},
		{
			name: "no content type",
			r: &http.Response{
				Header: http.Header{},
			},
			want: false,
		},
		{
			name: "html response with charset",
			r: &http.Response{
				Header: http.Header{"Content-Type": {"text/html; charset=utf-8"}},
			},
			want: true,
		},
		{
			name: "html response with charset and chunked encoding",
			r: &http.Response{
				Header: http.Header{"Content-Type": {"text/html; charset=utf-8"}, "Transfer-Encoding": {"chunked"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HtmlTransformCheck(tt.r)

			assert.Equal(t, tt.want, got)
		})
	}
}
