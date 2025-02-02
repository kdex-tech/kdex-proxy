package authn

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
			got := v.Validate(tt.args.r)
			assert.Equal(t, tt.wantErr, got != nil)
		})
	}
}

func TestNewAuthnMiddlewareFromEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want *AuthnMiddleware
	}{
		{
			name: "no protected paths",
			env:  "",
			want: &AuthnMiddleware{
				ProtectedPaths: []string{},
			},
		},
		{
			name: "protected paths",
			env:  "/protected",
			want: &AuthnMiddleware{
				ProtectedPaths: []string{"/protected"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(ProtectedPathsEnvVar, tt.env)
			got := NewAuthnMiddlewareFromEnv()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthnMiddleware_Authn(t *testing.T) {
	type fields struct {
		ProtectedPaths []string
		AuthValidators []iauthn.AuthValidator
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
			name: "no protected paths",
			fields: fields{
				ProtectedPaths: []string{},
			},
			args: args{
				h: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: "GET",
				path:   "/",
			},
		},
		{
			name: "protected path without auth",
			fields: fields{
				ProtectedPaths: []string{"/protected"},
				AuthValidators: []iauthn.AuthValidator{&iauthn.StaticBasicAuthValidator{Username: "testuser", Password: "testpassword"}},
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
				ProtectedPaths: []string{"/protected"},
				AuthValidators: []iauthn.AuthValidator{&iauthn.StaticBasicAuthValidator{Username: "testuser", Password: "testpassword"}},
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
				ProtectedPaths: []string{"/protected"},
				AuthValidators: []iauthn.AuthValidator{&iauthn.StaticBasicAuthValidator{
					Username:            "testuser",
					Password:            "testpassword",
					AuthorizationHeader: "Authorization",
					Realm:               "testrealm",
				}},
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
				ProtectedPaths: []string{"/protected"},
				AuthValidators: []iauthn.AuthValidator{&iauthn.StaticBasicAuthValidator{Username: "testuser", Password: "testpassword"}},
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
				ProtectedPaths: tt.fields.ProtectedPaths,
				AuthValidators: tt.fields.AuthValidators,
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
		ProtectedPaths []string
		AuthValidators []iauthn.AuthValidator
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []iauthn.AuthChallenge
	}{
		{
			name: "no protected paths",
			fields: fields{
				ProtectedPaths: []string{},
			},
			args: args{
				r: &http.Request{
					URL: &url.URL{Path: "/"},
				},
			},
			want: nil,
		},
		{
			name: "protected path without auth and no validators",
			fields: fields{
				ProtectedPaths: []string{"/protected"},
			},
			args: args{
				r: &http.Request{
					URL: &url.URL{Path: "/protected"},
				},
			},
			want: nil,
		},
		{
			name: "protected path without auth and with validators",
			fields: fields{
				ProtectedPaths: []string{"/protected"},
				AuthValidators: []iauthn.AuthValidator{&iauthn.StaticBasicAuthValidator{Username: "testuser", Password: "testpassword", Realm: "testrealm"}},
			},
			args: args{
				r: &http.Request{
					URL: &url.URL{Path: "/protected"},
				},
			},
			want: []iauthn.AuthChallenge{
				{
					Type:  "Basic",
					Realm: "testrealm",
				},
			},
		},
		{
			name: "protected path with invalid auth",
			fields: fields{
				ProtectedPaths: []string{"/protected"},
				AuthValidators: []iauthn.AuthValidator{&iauthn.StaticBasicAuthValidator{
					Username:            "testuser",
					Password:            "testpassword",
					AuthorizationHeader: "Authorization",
					Realm:               "testrealm",
				}},
			},
			args: args{
				r: &http.Request{
					URL:    &url.URL{Path: "/protected"},
					Header: http.Header{"Authorization": []string{"Basic bad"}},
				},
			},
			want: []iauthn.AuthChallenge{
				{
					Type:  "Basic",
					Realm: "testrealm",
				},
			},
		},
		{
			name: "protected path with valid auth",
			fields: fields{
				ProtectedPaths: []string{"/protected"},
				AuthValidators: []iauthn.AuthValidator{&iauthn.StaticBasicAuthValidator{
					AuthorizationHeader: "Authorization",
					Username:            "testuser",
					Password:            "testpassword",
					Realm:               "testrealm",
				}},
			},
			args: args{
				r: &http.Request{
					URL:    &url.URL{Path: "/protected"},
					Header: http.Header{"Authorization": []string{"Basic dGVzdHVzZXI6dGVzdHBhc3N3b3Jk"}},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthnMiddleware{
				ProtectedPaths: tt.fields.ProtectedPaths,
				AuthValidators: tt.fields.AuthValidators,
			}
			got := a.IsProtected(tt.args.r)
			assert.Equal(t, tt.want, got)
		})
	}
}
