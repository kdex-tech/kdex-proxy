package authz

import (
	"log"
	"net/http"

	khttp "kdex.dev/proxy/internal/http"
)

const ContextUserRolesKey khttp.ContextKey = "user_roles"

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
	pathPerms, err := a.permissionProvider.GetPermissions(r.URL.Path)
	if err != nil {
		// Don't accidentally give access on error
		log.Printf("Error getting permissions for path %s: %v", r.URL.Path, err)
		return err
	}

	if len(pathPerms) == 0 {
		// No permissions for this path, so we can give access by returning nil
		return nil
	}

	userRoles, ok := r.Context().Value(ContextUserRolesKey).([]string)
	if !ok {
		// No roles for this user, so we can't check permissions
		return ErrNoRoles
	}

	for _, perm := range pathPerms {
		if perm.Action == "view" && hasIntersection(perm.Principal, userRoles) {
			return nil
		}
	}

	// No permissions for this path, so we can't give access
	return ErrUnauthorized
}

func hasIntersection(a string, b []string) bool {
	for _, item := range b {
		if item == a {
			return true
		}
	}
	return false
}
