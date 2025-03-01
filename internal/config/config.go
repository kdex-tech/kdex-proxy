package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"hash/crc32"

	"gopkg.in/yaml.v3"
	"kdex.dev/proxy/internal/util"
)

type Config struct {
	Apps          []App             `json:"apps,omitempty" yaml:"apps,omitempty"`
	Authn         AuthnConfig       `json:"authn,omitempty" yaml:"authn,omitempty"`
	Authz         AuthzConfig       `json:"authz,omitempty" yaml:"authz,omitempty"`
	Expressions   ExpressionsConfig `json:"expressions,omitempty" yaml:"expressions,omitempty"`
	Fileserver    FileserverConfig  `json:"fileserver,omitempty" yaml:"fileserver,omitempty"`
	Importmap     ImportmapConfig   `json:"importmap,omitempty" yaml:"importmap,omitempty"`
	ListenAddress string            `json:"listen_address,omitempty" yaml:"listen_address,omitempty"`
	ListenPort    string            `json:"listen_port,omitempty" yaml:"listen_port,omitempty"`
	ModuleDir     string            `json:"module_dir,omitempty" yaml:"module_dir,omitempty"`
	Navigation    NavigationConfig  `json:"navigation,omitempty" yaml:"navigation,omitempty"`
	Proxy         ProxyConfig       `json:"proxy" yaml:"proxy"`
	Session       SessionConfig     `json:"session,omitempty" yaml:"session,omitempty"`
	State         StateConfig       `json:"state,omitempty" yaml:"state,omitempty"`
	hash          uint32
	json          bool
}

type App struct {
	Alias          string   `json:"alias,omitempty" yaml:"alias,omitempty"`
	Address        string   `json:"address" yaml:"address"`
	Element        string   `json:"element" yaml:"element"`
	Path           string   `json:"path" yaml:"path"`
	Targets        []Target `json:"targets" yaml:"targets"`
	RequiredScopes []string `json:"required_scopes,omitempty" yaml:"required_scopes,omitempty"`
}

type AuthnConfig struct {
	AuthenticateHeader     string          `json:"authenticate_header,omitempty" yaml:"authenticate_header,omitempty"`
	AuthorizationHeader    string          `json:"authorization_header,omitempty" yaml:"authorization_header,omitempty"`
	AuthenticateStatusCode int             `json:"authenticate_status_code,omitempty" yaml:"authenticate_status_code,omitempty"`
	AuthValidator          string          `json:"auth_validator,omitempty" yaml:"auth_validator,omitempty"`
	BasicAuth              BasicAuthConfig `json:"basic_auth,omitempty" yaml:"basic_auth,omitempty"`
	Login                  LoginConfig     `json:"login,omitempty" yaml:"login,omitempty"`
	Logout                 LogoutConfig    `json:"logout,omitempty" yaml:"logout,omitempty"`
	OAuth                  OAuthConfig     `json:"oauth,omitempty" yaml:"oauth,omitempty"`
	Realm                  string          `json:"realm,omitempty" yaml:"realm,omitempty"`
}

type AuthzConfig struct {
	CheckPrefix string                    `json:"check_prefix,omitempty" yaml:"check_prefix,omitempty"`
	Provider    string                    `json:"provider" yaml:"provider"`
	Static      StaticAuthzProviderConfig `json:"static,omitempty" yaml:"static,omitempty"`
}

type BasicAuthConfig struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

type ExpressionsConfig struct {
	Roles     string `json:"roles,omitempty" yaml:"roles,omitempty"`
	Principal string `json:"principal,omitempty" yaml:"principal,omitempty"`
}

type FileserverConfig struct {
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
}

type ImportmapConfig struct {
	PreloadModules []string `json:"preload_modules,omitempty" yaml:"preload_modules,omitempty"`
}

type LoginConfig struct {
	Path  string `json:"path" yaml:"path"`
	Label string `json:"label" yaml:"label"`
	Query string `json:"query" yaml:"query"`
}

type LogoutConfig struct {
	Path  string `json:"path" yaml:"path"`
	Label string `json:"label" yaml:"label"`
	Query string `json:"query" yaml:"query"`
}

type NavigationConfig struct {
	NavItemsQuery   string            `json:"nav_items_query" yaml:"nav_items_query"`
	NavItemFields   map[string]string `json:"nav_item_fields" yaml:"nav_item_fields"`
	NavItemTemplate string            `json:"nav_item_template" yaml:"nav_item_template"`
	ProtectedPaths  []string          `json:"protected_paths" yaml:"protected_paths"`
	TemplatePaths   []TemplatePath    `json:"template_paths" yaml:"template_paths"`
}

type OAuthConfig struct {
	AuthServerURL     string   `json:"auth_server_url" yaml:"auth_server_url"`
	ClientID          string   `json:"client_id" yaml:"client_id"`
	ClientSecret      string   `json:"client_secret" yaml:"client_secret"`
	DumpClaims        bool     `json:"dump_claims,omitempty" yaml:"dump_claims,omitempty"`
	Prefix            string   `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	RedirectURI       string   `json:"redirect_uri" yaml:"redirect_uri"`
	Scopes            []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	SignInOnChallenge bool     `json:"sign_in_on_challenge,omitempty" yaml:"sign_in_on_challenge,omitempty"`
}

type Permission struct {
	Action    string `json:"action" yaml:"action"`       // Action being performed (e.g., "read", "write")
	Principal string `json:"principal" yaml:"principal"` // Principal that can perform this action
	Resource  string `json:"resource" yaml:"resource"`   // Resource being accessed (e.g., "page", "api")
}

type ProxyConfig struct {
	AlwaysAppendSlash   bool   `json:"always_append_slash,omitempty" yaml:"always_append_slash,omitempty"`
	AppendIndex         bool   `json:"append_index,omitempty" yaml:"append_index,omitempty"`
	IndexFile           string `json:"index_file,omitempty" yaml:"index_file,omitempty"`
	PathSeparator       string `json:"path_separator,omitempty" yaml:"path_separator,omitempty"`
	ProbePath           string `json:"probe_path,omitempty" yaml:"probe_path,omitempty"`
	UpstreamAddress     string `json:"upstream_address" yaml:"upstream_address"`
	UpstreamPrefix      string `json:"upstream_prefix,omitempty" yaml:"upstream_prefix,omitempty"`
	UpstreamScheme      string `json:"upstream_scheme,omitempty" yaml:"upstream_scheme,omitempty"`
	UpstreamHealthzPath string `json:"upstream_healthz_path,omitempty" yaml:"upstream_healthz_path,omitempty"`
}

type SessionConfig struct {
	CookieName string `json:"cookie_name,omitempty" yaml:"cookie_name,omitempty"`
	Store      string `json:"store,omitempty" yaml:"store,omitempty"`
}

type StateConfig struct {
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
}

type StaticAuthzProviderConfig struct {
	Permissions []Permission `json:"permissions,omitempty" yaml:"permissions,omitempty"`
}

type Target struct {
	Path      string `json:"path" yaml:"path"`
	Container string `json:"container_id,omitempty" yaml:"container_id,omitempty"`
}

type TemplatePath struct {
	Href     string  `json:"href" yaml:"href"`
	Label    string  `json:"label" yaml:"label"`
	Template string  `json:"template" yaml:"template"`
	Weight   float64 `json:"weight" yaml:"weight"`
}

type RolesConfig struct {
	Expression string `json:"expression" yaml:"expression"`
}

var defaultConfig = Config{
	Authn: AuthnConfig{
		AuthenticateHeader:     "WWW-Authenticate",
		AuthorizationHeader:    "Authorization",
		AuthenticateStatusCode: 401,
		AuthValidator:          "static_basic_auth",
		BasicAuth: BasicAuthConfig{
			Username: "admin",
			Password: "admin",
		},
		Login: LoginConfig{
			Path:  "/~/oauth/login",
			Label: "Login",
		},
		Logout: LogoutConfig{
			Path:  "/~/oauth/logout",
			Label: "Logout",
		},
		OAuth: OAuthConfig{
			ClientID:    "kdex-proxy",
			Prefix:      "/~/oauth",
			RedirectURI: "/~/oauth/callback",
		},
		Realm: "KDEX Proxy",
	},
	Authz: AuthzConfig{
		CheckPrefix: "/~/check",
		Provider:    "static",
	},
	Expressions: ExpressionsConfig{
		Principal: "data.preferred_username",
		Roles:     "data.roles",
	},
	Fileserver: FileserverConfig{
		Prefix: "/~/m/",
	},
	Importmap: ImportmapConfig{
		PreloadModules: []string{
			"@kdex/ui",
		},
	},
	ListenAddress: "",
	ListenPort:    "8080",
	ModuleDir:     "/modules",
	Navigation: NavigationConfig{
		NavItemsQuery:   `nav`,
		NavItemFields:   map[string]string{},
		NavItemTemplate: ``,
		ProtectedPaths:  []string{},
		TemplatePaths:   []TemplatePath{},
	},
	Proxy: ProxyConfig{
		AlwaysAppendSlash:   false,
		AppendIndex:         false,
		IndexFile:           "index.html",
		PathSeparator:       "/_/",
		ProbePath:           "/~/p/{$}",
		UpstreamScheme:      "http",
		UpstreamHealthzPath: "/",
	},
	Session: SessionConfig{
		CookieName: "session_id",
		Store:      "memory",
	},
	State: StateConfig{
		Endpoint: "/~/state",
	},
}

func DefaultConfig() Config {
	return defaultConfig
}

func NewConfigFromEnv() Config {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "/etc/kdex-proxy/proxy.config"
	}

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("Error reading config file %s: %v", configFile, err)
		return defaultConfig
	}

	config := defaultConfig

	x := bytes.TrimLeft(configBytes, " \t\r\n")
	isJsonObject := len(x) > 0 && x[0] == '{'

	if isJsonObject {
		config.json = true
		err = json.Unmarshal(configBytes, &config)
		if err != nil {
			log.Printf("Error unmarshalling config: %v", err)
			return defaultConfig
		}
	} else {
		config.json = false
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			log.Printf("Error unmarshalling config: %v", err)
			return defaultConfig
		}
	}

	for _, app := range config.Apps {
		if err := app.Validate(); err != nil {
			log.Fatalf("Invalid app: %v", err)
		}
	}

	config.prettyPrint()

	return config
}

func (a *App) Validate() error {
	if a.Address == "" {
		return fmt.Errorf("app address is required")
	}
	if a.Element == "" {
		return fmt.Errorf("app element is required")
	}
	if a.Path == "" {
		return fmt.Errorf("app path is required")
	}
	if len(a.Targets) == 0 {
		return fmt.Errorf("app must have at least one target")
	}
	for _, target := range a.Targets {
		if target.Path == "" {
			return fmt.Errorf("app targets page is required")
		}
	}

	if a.Alias == "" {
		a.Alias = util.RandStringBytes(4)
	}

	return nil
}

func (c *Config) GetAppsForTargetPath(targetPath string) []App {
	var filteredApps []App
	for _, app := range c.Apps {
		for _, appTarget := range app.Targets {
			if appTarget.Path == targetPath {
				filteredApps = append(filteredApps, app)
			}
		}
	}
	return filteredApps
}

func (c *Config) prettyPrint() {
	var s []byte
	if c.json {
		s, _ = json.MarshalIndent(c, "", "  ")
	} else {
		s, _ = yaml.Marshal(c)
	}
	log.Printf("Using config:\n%s", string(s))
}

func (c *Config) Hash() uint32 {
	if c.hash != 0 {
		return c.hash
	}
	configBytes, err := json.Marshal(c)
	if err != nil {
		return 0
	}
	c.hash = crc32.ChecksumIEEE(configBytes)
	return c.hash
}
