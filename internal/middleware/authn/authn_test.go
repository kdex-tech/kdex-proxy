package authn

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	iauthn "kdex.dev/proxy/internal/authn"
)

func TestStaticBasicAuthValidator_Validate(t *testing.T) {
	type fields struct {
		Username string
		Password string
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid basic auth",
			fields: fields{
				Username: "testuser",
				Password: "testpassword",
			},
			args: args{
				r: &http.Request{
					Header: http.Header{"Authorization": []string{"Basic dGVzdHVzZXI6dGVzdHBhc3N3b3Jk"}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid basic auth",
			fields: fields{
				Username: "testuser",
				Password: "testpassword",
			},
			args: args{
				r: &http.Request{
					Header: http.Header{"Authorization": []string{"Basic bad"}},
				},
			},
			wantErr: true,
		},
		{
			name: "no basic auth",
			fields: fields{
				Username: "testuser",
				Password: "testpassword",
			},
			args: args{
				r: &http.Request{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &iauthn.StaticBasicAuthValidator{
				AuthorizationHeader: "Authorization",
				Username:            tt.fields.Username,
				Password:            tt.fields.Password,
			}
			got := v.Validate(httptest.NewRecorder(), tt.args.r)
			assert.Equal(t, tt.wantErr, got != nil)
		})
	}
}

func TestAuthnMiddleware_Authn(t *testing.T) {
	type fields struct {
		AuthValidator iauthn.AuthValidator
	}
	type args struct {
		h        http.Handler
		method   string
		path     string
		username string
		password string
		wantErr  bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "protected path without auth",
			fields: fields{
				AuthValidator: &iauthn.StaticBasicAuthValidator{
					AuthenticateHeader:     "WWW-Authenticate",
					AuthenticateStatusCode: http.StatusUnauthorized,
					AuthorizationHeader:    "Authorization",
					Username:               "testuser",
					Password:               "testpassword",
					Realm:                  "testrealm",
				},
			},
			args: args{
				h: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method:  "GET",
				path:    "/protected",
				wantErr: true,
			},
		},
		{
			name: "protected path with invalid auth",
			fields: fields{
				AuthValidator: &iauthn.StaticBasicAuthValidator{
					AuthenticateHeader:     "WWW-Authenticate",
					AuthenticateStatusCode: http.StatusUnauthorized,
					AuthorizationHeader:    "Authorization",
					Username:               "testuser",
					Password:               "testpassword",
					Realm:                  "testrealm",
				},
			},
			args: args{
				h: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method:   "GET",
				path:     "/protected",
				username: "testuser",
				password: "badpassword",
				wantErr:  true,
			},
		},
		{
			name: "protected path with valid auth",
			fields: fields{
				AuthValidator: &iauthn.StaticBasicAuthValidator{
					AuthenticateHeader:     "WWW-Authenticate",
					AuthenticateStatusCode: http.StatusUnauthorized,
					AuthorizationHeader:    "Authorization",
					Username:               "testuser",
					Password:               "testpassword",
					Realm:                  "testrealm",
				},
			},
			args: args{
				h: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method:   "GET",
				path:     "/protected",
				username: "testuser",
				password: "testpassword",
				wantErr:  false,
			},
		},
		{
			name: "not a protected path	",
			fields: fields{
				AuthValidator: &iauthn.StaticBasicAuthValidator{
					AuthenticateHeader:     "WWW-Authenticate",
					AuthenticateStatusCode: http.StatusUnauthorized,
					AuthorizationHeader:    "Authorization",
					Username:               "testuser",
					Password:               "testpassword",
					Realm:                  "testrealm",
				},
			},
			args: args{
				h: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method:   "GET",
				path:     "/notprotected",
				username: "testuser",
				password: "testpassword",
				wantErr:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthnMiddleware{
				AuthenticateStatusCode: http.StatusUnauthorized,
				AuthValidator:          tt.fields.AuthValidator,
			}
			got := a.Authn(tt.args.h)
			if got == nil {
				t.Errorf("AuthnMiddleware.Authn() = %v", got)
				return
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			if tt.args.username != "" {
				request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(tt.args.username+":"+tt.args.password)))
			}
			got.ServeHTTP(recorder, request)
			if tt.args.wantErr {
				if recorder.Code != http.StatusUnauthorized {
					t.Errorf("AuthnMiddleware.Authn() = %v, want %v", recorder.Code, http.StatusUnauthorized)
				}
			} else {
				if recorder.Code != http.StatusOK {
					t.Errorf("AuthnMiddleware.Authn() = %v, want %v", recorder.Code, http.StatusOK)
				}
			}
		})
	}
}

func TestAuthnMiddleware_IsProtected(t *testing.T) {
	type fields struct {
		AuthValidator iauthn.AuthValidator
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "protected path without auth and no validators",
			fields: fields{
				AuthValidator: &iauthn.NoOpAuthValidator{},
			},
			args: args{
				r: &http.Request{
					URL: &url.URL{Path: "/protected"},
				},
			},
			want: false,
		},
		{
			name: "protected path without auth and with validators",
			fields: fields{
				AuthValidator: &iauthn.StaticBasicAuthValidator{
					AuthenticateHeader:     "WWW-Authenticate",
					AuthenticateStatusCode: http.StatusUnauthorized,
					AuthorizationHeader:    "Authorization",
					Username:               "testuser",
					Password:               "testpassword",
					Realm:                  "testrealm",
				},
			},
			args: args{
				r: &http.Request{
					URL: &url.URL{Path: "/protected"},
				},
			},
			want: true,
		},
		{
			name: "protected path with invalid auth",
			fields: fields{
				AuthValidator: &iauthn.StaticBasicAuthValidator{
					AuthenticateHeader:     "WWW-Authenticate",
					AuthenticateStatusCode: http.StatusUnauthorized,
					AuthorizationHeader:    "Authorization",
					Username:               "testuser",
					Password:               "testpassword",
					Realm:                  "testrealm",
				},
			},
			args: args{
				r: &http.Request{
					URL:    &url.URL{Path: "/protected"},
					Header: http.Header{"Authorization": []string{"Basic bad"}},
				},
			},
			want: true,
		},
		{
			name: "protected path with valid auth",
			fields: fields{
				AuthValidator: &iauthn.StaticBasicAuthValidator{
					AuthenticateHeader:     "WWW-Authenticate",
					AuthenticateStatusCode: http.StatusUnauthorized,
					AuthorizationHeader:    "Authorization",
					Username:               "testuser",
					Password:               "testpassword",
					Realm:                  "testrealm",
				},
			},
			args: args{
				r: &http.Request{
					URL:    &url.URL{Path: "/protected"},
					Header: http.Header{"Authorization": []string{"Basic dGVzdHVzZXI6dGVzdHBhc3N3b3Jk"}},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthnMiddleware{
				AuthenticateStatusCode: http.StatusUnauthorized,
				AuthValidator:          tt.fields.AuthValidator,
			}
			got := a.IsProtected(httptest.NewRecorder(), tt.args.r)
			if got == nil && tt.want {
				t.Errorf("AuthnMiddleware.IsProtected() = nil, want %v", tt.want)
			}
			if got != nil && !tt.want {
				t.Errorf("AuthnMiddleware.IsProtected() != nil, want %v", tt.want)
			}
		})
	}
}
