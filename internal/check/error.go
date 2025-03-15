package check

import "errors"

var (
	ErrNoRoles = errors.New("no roles found in request context")
)
