package authn

import (
	"context"
	"log"
	"net/http"

	"kdex.dev/proxy/internal/config"
	khttp "kdex.dev/proxy/internal/http"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/store/state"
)

const (
	AuthScheme_Basic  = "Basic"
	AuthScheme_Bearer = "Bearer"

	Validator_NoOp            = "noop"
	Validator_StaticBasicAuth = "static_basic_auth"
	Validator_OAuth           = "oauth"
)

const ContextUserKey khttp.ContextKey = "user"

type AuthValidator interface {
	Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler)
	Register(mux *http.ServeMux)
}

func AuthValidatorFactory(
	config *config.AuthnConfig,
	sessionStore *session.SessionStore,
	sessionCookieName string,
) AuthValidator {
	var auth_validator AuthValidator
	switch config.AuthValidator {
	case Validator_StaticBasicAuth:
		auth_validator = &StaticBasicAuthValidator{
			AuthorizationHeader:    config.AuthorizationHeader,
			AuthenticateHeader:     config.AuthenticateHeader,
			AuthenticateStatusCode: config.AuthenticateStatusCode,
			Realm:                  config.Realm,
			Username:               config.BasicAuth.Username,
			Password:               config.BasicAuth.Password,
		}
	case Validator_OAuth:
		stateStore, err := state.NewStateStore(context.Background(), "memory")
		if err != nil {
			log.Fatalf("Failed to create state store: %v", err)
		}
		auth_validator = NewOAuthValidator(
			context.Background(),
			config,
			sessionCookieName,
			sessionStore,
			&stateStore,
		)
	default: // Validator_NoOp
		auth_validator = &NoOpAuthValidator{}
	}

	return auth_validator
}
