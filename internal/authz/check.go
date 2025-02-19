package authz

import (
	"context"
	"errors"
)

type Checker struct {
	PermissionProvider PermissionProvider
}

func (c *Checker) Check(ctx context.Context, resource string, action string) (bool, error) {
	perms, err := c.PermissionProvider.GetPermissions(resource)
	if errors.Is(err, ErrNoPermissions) {
		return false, err
	}

	userRoles, ok := ctx.Value(ContextUserRolesKey).([]string)
	if !ok {
		return false, ErrNoRoles
	}

	for _, perm := range perms {
		if (perm.Action == "*" || perm.Action == action) && c.hasIntersection(perm.Principal, userRoles) {
			return true, nil
		}
	}

	return false, nil
}

func (c *Checker) hasIntersection(a string, b []string) bool {
	for _, item := range b {
		if item == a {
			return true
		}
	}
	return false
}
