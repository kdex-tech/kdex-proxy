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

package state

import (
	"encoding/json"
	"log"
	"net/http"

	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/store/session"
)

type UserState struct {
	Principal string                 `json:"principal"`
	Roles     []string               `json:"roles"`
	Data      map[string]interface{} `json:"data"`
}

type StateHandler struct {
	FieldEvaluator *expression.FieldEvaluator
}

func (h *StateHandler) StateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sessionData, ok := r.Context().Value(kctx.SessionDataKey).(*session.SessionData)
		if !ok || sessionData == nil {
			json.NewEncoder(w).Encode(UserState{
				Principal: "",
				Roles:     []string{},
				Data:      map[string]interface{}{},
			})
			return
		}

		principal, err := h.FieldEvaluator.EvaluatePrincipal(sessionData.Data)
		if err != nil {
			log.Printf("error evaluating principal: %v", err)
			principal = ""
		}

		roles, err := h.FieldEvaluator.EvaluateRoles(sessionData.Data)
		if err != nil {
			log.Printf("error evaluating roles: %v", err)
			roles = []string{}
		}

		json.NewEncoder(w).Encode(UserState{
			Principal: principal,
			Roles:     roles,
			Data:      sessionData.Data,
		})
	}
}
