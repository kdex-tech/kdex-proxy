package authn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpAuthValidator_Register(t *testing.T) {
	tests := []struct {
		name string
		v    *NoOpAuthValidator
	}{
		{
			name: "register",
			v:    &NoOpAuthValidator{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.v
			v.Register(nil)
		})
	}
}

func TestNoOpAuthValidator_Validate(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		v    *NoOpAuthValidator
		args args
		want func(h http.Handler)
	}{
		{
			name: "validate",
			v:    &NoOpAuthValidator{},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.v
			got := v.Validate(tt.args.w, tt.args.r)
			assert.True(t, tt.want == nil, got == nil)
		})
	}
}
