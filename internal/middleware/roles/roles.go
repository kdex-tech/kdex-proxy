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

package roles

import (
	"context"
	"log"
	"net/http"

	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/store/session"
)

type RolesMiddleware struct {
	FieldEvaluator *expression.FieldEvaluator
}

func (m *RolesMiddleware) InjectRoles(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var roles []string

		sessionData, ok := r.Context().Value(kctx.SessionDataKey).(*session.SessionData)
		if ok && sessionData != nil {
			var err error
			roles, err = m.FieldEvaluator.EvaluateRoles(sessionData.Data)
			if err != nil {
				log.Printf("failed to evaluate roles: %v", err)
				roles = []string{"anonymous"}
			}
		} else {
			roles = []string{"anonymous"}
		}

		// Add roles to context
		ctx := context.WithValue(r.Context(), kctx.UserRolesKey, roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
