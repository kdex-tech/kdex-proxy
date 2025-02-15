package authn

import (
	"encoding/json"
	"net/http"

	"kdex.dev/proxy/internal/roles"
	"kdex.dev/proxy/internal/store/session"
)

type UserState struct {
	IsLoggedIn bool                   `json:"isLoggedIn"`
	Roles      []string               `json:"roles"`
	Data       map[string]interface{} `json:"data"`
}

type StateHandler struct {
	RoleEvaluator *roles.RoleEvaluator
}

func (h *StateHandler) StateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sessionData, ok := r.Context().Value(ContextUserKey).(*session.SessionData)
		if !ok {
			json.NewEncoder(w).Encode(UserState{
				IsLoggedIn: false,
				Roles:      []string{},
				Data:       map[string]interface{}{},
			})
			return
		}

		roles, err := h.RoleEvaluator.EvaluateRoles(sessionData.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(UserState{
			IsLoggedIn: true,
			Roles:      roles,
			Data:       sessionData.Data,
		})
	}
}
