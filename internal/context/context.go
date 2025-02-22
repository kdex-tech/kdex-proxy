package context

type ContextKey string

const (
	ProxiedPartsKey ContextKey = "proxiedParts"
	SessionDataKey  ContextKey = "sessionData"
	UserRolesKey    ContextKey = "userRoles"
)

type ProxiedParts struct {
	AppAlias    string
	AppPath     string
	ProxiedPath string
}
