package authz

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized access")
)
