package authz

import (
	"strings"

	"kdex.dev/proxy/internal/config"
)

// PermissionProvider defines the interface for retrieving permissions
type PermissionProvider interface {
	GetPermissions(resource string) ([]config.Permission, error)
}

func NewPermissionProvider(config *config.Config) PermissionProvider {
	if config.Authz.Provider == "static" {
		return &StaticPermissionProvider{
			Permissions: config.Authz.Static.Permissions,
		}
	}

	return nil
}

// StaticPermissionProvider implements PermissionProvider with a static map
type StaticPermissionProvider struct {
	Permissions []config.Permission
}

func (p *StaticPermissionProvider) GetPermissions(resource string) ([]config.Permission, error) {
	parts := strings.SplitN(resource, ":", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidResource
	}

	resourceType := parts[0]
	resourceKey := parts[1]

	filtered := []config.Permission{}
	for _, perm := range p.Permissions {
		_parts := strings.SplitN(perm.Resource, ":", 2)

		if len(_parts) != 2 {
			return nil, ErrInvalidResource
		}

		_resourceType := _parts[0]
		_resourceKey := _parts[1]

		if (_resourceType == "*" || _resourceType == resourceType) &&
			(_resourceKey == "*" || _resourceKey == resourceKey ||
				(strings.HasSuffix(_resourceKey, "*") &&
					strings.HasPrefix(resourceKey, _resourceKey[:len(_resourceKey)-1]))) {

			filtered = append(filtered, perm)
		}
	}
	if len(filtered) == 0 {
		return nil, ErrNoPermissions
	}
	return filtered, nil
}
