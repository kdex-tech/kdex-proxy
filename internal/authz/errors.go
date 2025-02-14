package authz

import "errors"

var (
	ErrNoRoles      = errors.New("no roles found in request context")
	ErrUnauthorized = errors.New("unauthorized access")
)
