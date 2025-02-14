package authz

import (
	"net/http"
)

// Authorizer defines the interface for authorization checks
type Authorizer interface {
	// CheckAccess verifies if the request has permission to access the resource
	CheckAccess(r *http.Request) error
}

// NewAuthorizer creates a new Authorizer based on configuration
func NewAuthorizer(permissionProvider PermissionProvider) Authorizer {
	return &defaultAuthorizer{
		permissionProvider: permissionProvider,
	}
}

type defaultAuthorizer struct {
	permissionProvider PermissionProvider
}

func (a *defaultAuthorizer) CheckAccess(r *http.Request) error {
	userRoles, ok := r.Context().Value("user_roles").([]string)
	if !ok {
		return ErrNoRoles
	}

	pathPerms, err := a.permissionProvider.GetPermissions(r.URL.Path)
	if err != nil {
		return err
	}

	// Check path-specific permissions first
	for _, perm := range pathPerms {
		if !hasIntersection(perm.Roles, userRoles) {
			return ErrUnauthorized
		}
	}

	return nil
}

func hasIntersection(a, b []string) bool {
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}
	for _, item := range b {
		if set[item] {
			return true
		}
	}
	return false
}
