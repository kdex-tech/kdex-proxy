package permission

import "errors"

var (
	ErrInvalidResource = errors.New("invalid resource: format must be <type>:<action>")
	ErrNoPermissions   = errors.New("no permissions found for resource")
)
