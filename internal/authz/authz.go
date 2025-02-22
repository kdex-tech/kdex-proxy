package authz

import (
	"errors"
	"fmt"
	"net/http"
)

// Authorizer defines the interface for authorization checks
type Authorizer interface {
	// CheckAccess verifies if the request has permission to access the resource
	CheckAccess(r *http.Request) error
}

// NewAuthorizer creates a new Authorizer based on configuration
func NewAuthorizer(checker Checker) Authorizer {
	return &defaultAuthorizer{
		checker: checker,
	}
}

type defaultAuthorizer struct {
	checker Checker
}

func (a *defaultAuthorizer) CheckAccess(r *http.Request) error {
	ok, err := a.checker.Check(r.Context(), fmt.Sprintf("page:%s", r.URL.Path), "view")

	if err != nil && !errors.Is(err, ErrNoPermissions) && !errors.Is(err, ErrNoRoles) {
		return err
	}

	if ok {
		return nil
	}

	return ErrUnauthorized
}
