package check

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"kdex.dev/proxy/internal/authz"
)

type CheckHandler struct {
	Checker *authz.Checker
}

// func (v *CheckHandler) Register(mux *http.ServeMux) {
// 	mux.HandleFunc("GET "+v.Config.Check.Prefix+"/check", v.CheckHandler())
// }

func (h *CheckHandler) CheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		action := r.URL.Query()["action"]
		resource := r.URL.Query()["resource"]

		allowed, err := h.Checker.Check(r.Context(), resource[0], action[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"allowed": %t}`, allowed)))
	}
}

func (h *CheckHandler) CheckBatchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
			return
		}

		var requestBody struct {
			Tuples []authz.CheckBatchTuples `json:"tuples"`
		}

		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
			return
		}

		results, err := h.Checker.CheckBatch(r.Context(), requestBody.Tuples)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results)
	}
}
