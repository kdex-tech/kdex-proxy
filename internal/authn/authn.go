package authn

import (
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
	DefaultRealm                  = "KDEX Proxy"

	Validator_NoOp            = "noop"
	Validator_StaticBasicAuth = "static_basic_auth"
	Validator_BearerAuth      = "bearer_auth"
)

type AuthChallenge struct {
	Scheme     string
	Attributes map[string]string
}

type AuthValidator interface {
	Validate(r *http.Request) *AuthChallenge
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
			AuthorizationHeader: authorization_header,
			Realm:               realm,
			Username:            username,
			Password:            password,
		}
	case Validator_BearerAuth:
		realm := os.Getenv("BEARER_AUTH_REALM")
		if realm == "" {
			realm = DefaultRealm
		}
		scopes := os.Getenv("BEARER_AUTH_SCOPES")
		if scopes == "" {
			scopes = "read write"
		}
		auth_validator = &BearerAuthValidator{
			AuthorizationHeader: authorization_header,
			Realm:               realm,
			Scopes:              strings.Split(scopes, " "),
		}
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
