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

	iauthn "kdex.dev/proxy/internal/authn"
)

type AuthnMiddleware struct {
	AuthenticateHeader     string
	AuthenticateStatusCode int
	AuthValidator          iauthn.AuthValidator
}

func (a *AuthnMiddleware) Authn(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		challenge := a.IsProtected(w, r)

		if challenge != nil {
			challenge(h)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func (a *AuthnMiddleware) IsProtected(w http.ResponseWriter, r *http.Request) func(h http.Handler) {
	challenge := a.AuthValidator.Validate(w, r)
	if challenge != nil {
		return challenge
	}
	return nil
}
