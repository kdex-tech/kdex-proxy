package authn

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	AuthScheme_Basic  = "Basic"
	AuthScheme_Bearer = "Bearer"

	DefaultAuthenticateHeader     = "WWW-Authenticate"
	DefaultAuthorizationHeader    = "Authorization"
	DefaultAuthenticateStatusCode = 401
	DefaultPrefix                 = "/~/o/"
	DefaultRealm                  = "KDEX Proxy"

	Validator_NoOp            = "noop"
	Validator_StaticBasicAuth = "static_basic_auth"
	Validator_OAuth           = "oauth"
)

type ContextKey string

const ContextUserKey ContextKey = "user"

type AuthValidator interface {
	Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler)
	Register(mux *http.ServeMux)
}

type AuthnConfig struct {
	AuthenticateHeader     string
	AuthorizationHeader    string
	AuthenticateStatusCode int
	AuthValidators         []AuthValidator
	ProtectedPaths         []string
}

func NewAuthnConfigFromEnv() *AuthnConfig {
	authenticate_header := os.Getenv("AUTHENTICATE_HEADER")
	if authenticate_header == "" {
		authenticate_header = DefaultAuthenticateHeader
	}
	log.Printf("Authenticate header: %s", authenticate_header)

	authorization_header := os.Getenv("AUTHORIZATION_HEADER")
	if authorization_header == "" {
		authorization_header = DefaultAuthorizationHeader
	}
	log.Printf("Authorization header: %s", authorization_header)
	authenticate_status_code, err := strconv.Atoi(os.Getenv("AUTHENTICATE_STATUS_CODE"))
	if err != nil {
		authenticate_status_code = DefaultAuthenticateStatusCode
	}
	log.Printf("Authenticate status code: %d", authenticate_status_code)

	auth_validator_env := os.Getenv("AUTH_VALIDATOR")
	if auth_validator_env == "" {
		auth_validator_env = Validator_NoOp
	}
	log.Printf("Auth validator: %s", auth_validator_env)

	var protected_paths []string
	protected_paths_env := os.Getenv("PROTECTED_PATHS")
	if protected_paths_env == "" {
		protected_paths = []string{}
	} else {
		protected_paths = strings.Split(protected_paths_env, ",")
	}
	log.Printf("Protected paths: %v", protected_paths)

	var auth_validator AuthValidator
	switch auth_validator_env {
	case Validator_StaticBasicAuth:
		realm := os.Getenv("STATIC_BASIC_AUTH_REALM")
		if realm == "" {
			realm = DefaultRealm
		}
		username := os.Getenv("STATIC_BASIC_AUTH_USERNAME")
		password := os.Getenv("STATIC_BASIC_AUTH_PASSWORD")
		if username == "" || password == "" {
			log.Fatalf("STATIC_BASIC_AUTH_USERNAME and STATIC_BASIC_AUTH_PASSWORD must be set when using %s validator", Validator_StaticBasicAuth)
		}
		auth_validator = &StaticBasicAuthValidator{
			AuthorizationHeader:    authorization_header,
			AuthenticateHeader:     authenticate_header,
			AuthenticateStatusCode: authenticate_status_code,
			Realm:                  realm,
			Username:               username,
			Password:               password,
		}
	case Validator_OAuth:
		auth_server_url := os.Getenv("OAUTH_SERVER_URL")
		if auth_server_url == "" {
			log.Fatalf("OAUTH_SERVER_URL must be set when using %s validator", Validator_OAuth)
		}
		realm := os.Getenv("OAUTH_REALM")
		if realm == "" {
			realm = DefaultRealm
		}
		client_id := os.Getenv("OAUTH_CLIENT_ID")
		if client_id == "" {
			log.Fatalf("OAUTH_CLIENT_ID must be set when using %s validator", Validator_OAuth)
		}
		client_secret := os.Getenv("OAUTH_CLIENT_SECRET")
		if client_secret == "" {
			log.Fatalf("OAUTH_CLIENT_SECRET must be set when using %s validator", Validator_OAuth)
		}
		redirect_uri := os.Getenv("OAUTH_REDIRECT_URI")
		if redirect_uri == "" {
			log.Fatalf("OAUTH_REDIRECT_URI must be set when using %s validator", Validator_OAuth)
		}
		scopes := os.Getenv("OAUTH_SCOPES")
		if scopes == "" {
			scopes = "read write"
		}
		prefix := os.Getenv("OAUTH_PREFIX")
		if prefix == "" {
			prefix = DefaultPrefix
		}
		auth_validator = NewOAuthValidator(context.Background(), &Config{
			AuthorizationHeader: authorization_header,
			AuthServerURL:       auth_server_url,
			ClientID:            client_id,
			ClientSecret:        client_secret,
			Prefix:              prefix,
			Realm:               realm,
			RedirectURI:         redirect_uri,
			Scopes:              strings.Split(scopes, " "),
		})
	default: // Validator_NoOp
		auth_validator = &NoOpAuthValidator{}
	}

	return &AuthnConfig{
		AuthenticateHeader:     authenticate_header,
		AuthorizationHeader:    authorization_header,
		AuthenticateStatusCode: authenticate_status_code,
		AuthValidators:         []AuthValidator{auth_validator},
		ProtectedPaths:         protected_paths,
	}
}

func (c *AuthnConfig) Register(mux *http.ServeMux) *AuthnConfig {
	for _, validator := range c.AuthValidators {
		validator.Register(mux)
	}
	return c
}
