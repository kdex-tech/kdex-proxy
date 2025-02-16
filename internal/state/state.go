package state

import (
	"encoding/json"
	"log"
	"net/http"

	"kdex.dev/proxy/internal/authn"
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

		sessionData, ok := r.Context().Value(authn.ContextUserKey).(*session.SessionData)
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
