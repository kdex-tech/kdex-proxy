// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authn

import (
	"net/http"

	"kdex.dev/proxy/internal/config"
)

const (
	AuthScheme_Basic  = "Basic"
	AuthScheme_Bearer = "Bearer"

	Validator_NoOp            = "noop"
	Validator_StaticBasicAuth = "static_basic_auth"
	Validator_OAuth           = "oauth"
)

type AuthValidator interface {
	Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler)
	Register(mux *http.ServeMux)
}

func AuthValidatorFactory(config *config.Config) AuthValidator {
	var auth_validator AuthValidator

	switch config.Authn.AuthValidator {
	case Validator_StaticBasicAuth:
		auth_validator = &StaticBasicAuthValidator{
			AuthorizationHeader:    config.Authn.AuthorizationHeader,
			AuthenticateHeader:     config.Authn.AuthenticateHeader,
			AuthenticateStatusCode: config.Authn.AuthenticateStatusCode,
			Realm:                  config.Authn.Realm,
			Username:               config.Authn.BasicAuth.Username,
			Password:               config.Authn.BasicAuth.Password,
		}
	case Validator_OAuth:
		auth_validator = NewOAuthValidator(config)
	default: // Validator_NoOp
		auth_validator = &NoOpAuthValidator{}
	}

	return auth_validator
}
