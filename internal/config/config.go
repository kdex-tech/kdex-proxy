package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"kdex.dev/proxy/internal/util"
)

type Config struct {
	Apps          Apps             `json:"apps,omitempty"`
	Authn         AuthnConfig      `json:"authn"`
	Fileserver    FileserverConfig `json:"fileserver"`
	Importmap     ImportmapConfig  `json:"importmap"`
	ListenAddress string           `json:"listen_address"`
	ListenPort    string           `json:"listen_port"`
	ModuleDir     string           `json:"module_dir"`
	Proxy         ProxyConfig      `json:"proxy"`
	Session       SessionConfig    `json:"session"`
}

type Apps []App

type App struct {
	Alias          string   `json:"alias"`
	Address        string   `json:"address"`
	Element        string   `json:"element"`
	Path           string   `json:"path"`
	Targets        []Target `json:"targets"`
	RequiredScopes []string `json:"required_scopes,omitempty"`
}

type AuthnConfig struct {
	AuthenticateHeader     string          `json:"authenticate_header"`
	AuthorizationHeader    string          `json:"authorization_header"`
	AuthenticateStatusCode int             `json:"authenticate_status_code"`
	AuthValidator          string          `json:"auth_validator"`
	BasicAuth              BasicAuthConfig `json:"basic_auth"`
	OAuth                  OAuthConfig     `json:"oauth"`
	ProtectedPaths         []string        `json:"protected_paths"`
	Realm                  string          `json:"realm"`
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type FileserverConfig struct {
	Prefix string `json:"prefix"`
}

type ImportmapConfig struct {
	ModuleBody   string `json:"module_body"`
	ModulePrefix string `json:"module_prefix"`
}

type OAuthConfig struct {
	AuthServerURL     string   `json:"auth_server_url"`
	ClientID          string   `json:"client_id"`
	ClientSecret      string   `json:"client_secret"`
	DumpClaims        bool     `json:"dump_claims"`
	Prefix            string   `json:"prefix"`
	RedirectURI       string   `json:"redirect_uri"`
	Scopes            []string `json:"scopes"`
	SignInOnChallenge bool     `json:"sign_in_on_challenge,omitempty"`
}

type ProxyConfig struct {
	AlwaysAppendSlash   bool   `json:"always_append_slash"`
	PathSeparator       string `json:"path_separator"`
	ProbePath           string `json:"probe_path"`
	UpstreamAddress     string `json:"upstream_address"`
	UpstreamScheme      string `json:"upstream_scheme"`
	UpstreamHealthzPath string `json:"upstream_healthz_path"`
}

type SessionConfig struct {
	Store string `json:"store"`
}

type Target struct {
	Path      string `json:"path"`
	Container string `json:"container_id,omitempty"`
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
		OAuth: OAuthConfig{
			ClientID:    "kdex-proxy",
			Prefix:      "/~/o/",
			RedirectURI: "/~/o/oauth/callback",
			Scopes: []string{
				"read",
				"write",
			},
		},
		ProtectedPaths: []string{},
		Realm:          "KDEX Proxy",
	},
	Fileserver: FileserverConfig{
		Prefix: "/~/m/",
	},
	Importmap: ImportmapConfig{
		ModuleBody:   "import '@kdex/ui';",
		ModulePrefix: "/~/m/",
	},
	ListenAddress: "",
	ListenPort:    "8080",
	ModuleDir:     "/modules",
	Proxy: ProxyConfig{
		AlwaysAppendSlash:   false,
		PathSeparator:       "/_/",
		ProbePath:           "/~/p/{$}",
		UpstreamScheme:      "http",
		UpstreamHealthzPath: "/",
	},
	Session: SessionConfig{
		Store: "memory",
	},
}

func DefaultConfig() *Config {
	return &defaultConfig
}

func NewConfigFromEnv() *Config {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "/etc/kdex-proxy/config.json"
	}

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("Error reading config file %s: %v", configFile, err)
		return &defaultConfig
	}

	config := defaultConfig
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Printf("Error unmarshalling config: %v", err)
		return &defaultConfig
	}

	for _, app := range config.Apps {
		if err := app.Validate(); err != nil {
			log.Fatalf("Invalid app: %v", err)
		}
	}

	config.prettyPrint()

	return &config
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

func (a *Apps) GetAppsForTargetPath(targetPath string) []App {
	var filteredApps []App
	for _, app := range *a {
		for _, appTarget := range app.Targets {
			if appTarget.Path == targetPath {
				filteredApps = append(filteredApps, app)
			}
		}
	}
	return filteredApps
}

func (c *Config) prettyPrint() {
	s, _ := json.MarshalIndent(c, "", "  ")
	log.Printf("Using config: %s", string(s))
}
