package check

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/permission"
)

type Checker struct {
	PermissionProvider permission.PermissionProvider
}

func NewChecker(config *config.Config) *Checker {
	return &Checker{
		PermissionProvider: permission.NewPermissionProvider(config),
	}
}

func (c *Checker) SingleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		action := r.URL.Query()["action"]
		resource := r.URL.Query()["resource"]

		allowed, err := c.Check(r.Context(), resource[0], action[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"allowed": %t}`, allowed)))
	}
}

func (c *Checker) BatchHandler() http.HandlerFunc {
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
			Tuples []CheckBatchTuples `json:"tuples"`
		}

		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
			return
		}

		results, err := c.CheckBatch(r.Context(), requestBody.Tuples)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results)
	}
}

type CheckBatchResult struct {
	Resource string `json:"resource"`
	Allowed  bool   `json:"allowed"`
	Error    error  `json:"error"`
}

type CheckBatchTuples struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

func (c *Checker) Check(ctx context.Context, resource string, action string) (bool, error) {
	userRoles, ok := ctx.Value(kctx.UserRolesKey).([]string)
	if !ok || len(userRoles) == 0 {
		return false, ErrNoRoles
	}

	return c.check(userRoles, resource, action)
}

func (c *Checker) CheckBatch(ctx context.Context, tuples []CheckBatchTuples) ([]CheckBatchResult, error) {
	userRoles, ok := ctx.Value(kctx.UserRolesKey).([]string)
	if !ok || len(userRoles) == 0 {
		return nil, ErrNoRoles
	}

	results := make([]CheckBatchResult, len(tuples))
	for i, tuple := range tuples {
		allowed, err := c.check(userRoles, tuple.Resource, tuple.Action)
		if err != nil {
			results[i] = CheckBatchResult{Resource: tuple.Resource, Allowed: false, Error: err}
			continue
		}
		results[i] = CheckBatchResult{Resource: tuple.Resource, Allowed: allowed}
	}
	return results, nil
}

func (c *Checker) check(userRoles []string, resource string, action string) (bool, error) {
	perms, err := c.PermissionProvider.GetPermissions(resource)
	if errors.Is(err, permission.ErrNoPermissions) {
		return false, err
	}

	for _, perm := range perms {
		if perm.Action == "*" || perm.Action == action {
			for _, item := range userRoles {
				if strings.HasSuffix(perm.Principal, "*") {
					if strings.HasPrefix(item, perm.Principal[:len(perm.Principal)-1]) {
						return true, nil
					}
				} else {
					if item == perm.Principal {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}
