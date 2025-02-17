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
			permissions: config.Authz.Static.Permissions,
		}
	}

	return nil
}

// StaticPermissionProvider implements PermissionProvider with a static map
type StaticPermissionProvider struct {
	permissions []config.Permission
}

func (p *StaticPermissionProvider) GetPermissions(resource string) ([]config.Permission, error) {
	filtered := []config.Permission{}
	for _, perm := range p.permissions {
		if strings.HasPrefix(resource, perm.Resource) {
			filtered = append(filtered, perm)
		}
	}
	return filtered, nil
}
