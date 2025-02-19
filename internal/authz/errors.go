package authz

import "errors"

var (
	ErrInvalidResource = errors.New("invalid resource: format must be <type>:<action>")
	ErrNoPermissions   = errors.New("no permissions found for resource")
	ErrNoRoles         = errors.New("no roles found in request context")
	ErrUnauthorized    = errors.New("unauthorized access")
)
