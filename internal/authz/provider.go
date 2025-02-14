package authz

import "kdex.dev/proxy/internal/config"

// PermissionProvider defines the interface for retrieving permissions
type PermissionProvider interface {
	GetPermissions(path string) ([]config.Permission, error)
}

func NewPermissionProvider(config *config.Config) PermissionProvider {
	if config.Authz.Provider == "static" {
		return &StaticPermissionProvider{
			permissions: config.Authz.Static.Permissions,
		}
	}

	return nil
}

// StaticPermissionProvider implements PermissionProvider with a static map
type StaticPermissionProvider struct {
	permissions map[string][]config.Permission
}

func (p *StaticPermissionProvider) GetPermissions(path string) ([]config.Permission, error) {
	if perms, exists := p.permissions[path]; exists {
		return perms, nil
	}
	return []config.Permission{}, nil
}
