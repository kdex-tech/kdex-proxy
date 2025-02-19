package authz

import (
	"context"
	"errors"
	"strings"
)

type Checker struct {
	PermissionProvider PermissionProvider
}

func (c *Checker) Check(ctx context.Context, resource string, action string) (bool, error) {
	userRoles, ok := ctx.Value(ContextUserRolesKey).([]string)
	if !ok {
		return false, ErrNoRoles
	}

	return c.check(userRoles, resource, action)
}

type CheckBatchResult struct {
	Resource string
	Allowed  bool
	Error    error
}

func (c *Checker) CheckBatch(ctx context.Context, resources []string, action string) ([]CheckBatchResult, error) {
	userRoles, ok := ctx.Value(ContextUserRolesKey).([]string)
	if !ok {
		return nil, ErrNoRoles
	}

	results := make([]CheckBatchResult, len(resources))
	for i, resource := range resources {
		allowed, err := c.check(userRoles, resource, action)
		if err != nil {
			results[i] = CheckBatchResult{Resource: resource, Allowed: false, Error: err}
			continue
		}
		results[i] = CheckBatchResult{Resource: resource, Allowed: allowed}
	}
	return results, nil
}

func (c *Checker) check(userRoles []string, resource string, action string) (bool, error) {
	perms, err := c.PermissionProvider.GetPermissions(resource)
	if errors.Is(err, ErrNoPermissions) {
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
