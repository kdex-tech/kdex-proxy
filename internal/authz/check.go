package authz

import (
	"context"
	"errors"
	"strings"
)

type Checker struct {
	PermissionProvider PermissionProvider
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
	userRoles, ok := ctx.Value(ContextUserRolesKey).([]string)
	if !ok || len(userRoles) == 0 {
		return false, ErrNoRoles
	}

	return c.check(userRoles, resource, action)
}

func (c *Checker) CheckBatch(ctx context.Context, tuples []CheckBatchTuples) ([]CheckBatchResult, error) {
	userRoles, ok := ctx.Value(ContextUserRolesKey).([]string)
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
