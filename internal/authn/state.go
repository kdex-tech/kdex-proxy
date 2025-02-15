package authn

import (
	"encoding/json"
	"log"
	"net/http"

	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/store/session"
)

type UserState struct {
	Identity   string                 `json:"identity"`
	IsLoggedIn bool                   `json:"isLoggedIn"`
	Roles      []string               `json:"roles"`
	Data       map[string]interface{} `json:"data"`
}

type StateHandler struct {
	FieldEvaluator *expression.FieldEvaluator
}

func (h *StateHandler) StateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sessionData, ok := r.Context().Value(ContextUserKey).(*session.SessionData)
		if !ok || sessionData == nil {
			json.NewEncoder(w).Encode(UserState{
				Identity:   "",
				IsLoggedIn: false,
				Roles:      []string{},
				Data:       map[string]interface{}{},
			})
			return
		}

		identity, err := h.FieldEvaluator.EvaluateIdentity(sessionData.Data)
		if err != nil {
			log.Printf("error evaluating identity: %v", err)
			identity = ""
		}

		roles, err := h.FieldEvaluator.EvaluateRoles(sessionData.Data)
		if err != nil {
			log.Printf("error evaluating roles: %v", err)
			roles = []string{}
		}

		json.NewEncoder(w).Encode(UserState{
			Identity:   identity,
			IsLoggedIn: true,
			Roles:      roles,
			Data:       sessionData.Data,
		})
	}
}
