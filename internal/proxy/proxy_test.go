// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/app"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/importmap"
	"kdex.dev/proxy/internal/meta"
	"kdex.dev/proxy/internal/navigation"
	"kdex.dev/proxy/internal/transform"
)

type Result struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
}

// mockTargetServer creates a test server that simulates the target server
func mockTargetServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test-value")

		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}

		if r.URL.Path == "/test/html_without_importmap" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body><h1>Hello, World!</h1></body></html>`))
			return
		}

		if r.URL.Path == "/test/app1" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body><kdex-ui-app-container></kdex-ui-app-container></body></html>`))
			return
		}

		if r.URL.Path == "/test/html_with_importmap" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><script type="importmap">{"imports":{"@foo/bar":"/foo/bar.js"}}</script></head><body>test</body></html>`))
			return
		}

		if r.URL.Path == "/test/400" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if r.URL.Path == "/test/500" {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		result := Result{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: map[string][]string{},
		}

		for k, v := range r.Header {
			result.Headers[k] = v
		}

		bytes, _ := json.Marshal(result)

		fmt.Fprintf(w, "%s", string(bytes))
	}))
}

func TestServer_ReverseProxy(t *testing.T) {
	// Start mock target server
	targetServer := mockTargetServer()
	defer targetServer.Close()

	upstreamAddress := strings.TrimPrefix(targetServer.URL, "http://")
	defaultConfig := config.DefaultConfig()
	defaultConfig.Apps = []config.App{
		{
			Alias:   "app1",
			Address: "localhost:61345",
			Element: "app-one",
			Path:    "/app1.js",
			Targets: []config.Target{
				{
					Path:      "/test/app1",
					Container: "main",
				},
			},
		},
	}
	defaultConfig.Authn.Login.Query = `nav a[href="/signin/"]`
	defaultConfig.Authn.Login.Label = "Login"
	defaultConfig.Authn.Login.Path = "/~/oauth/login"
	defaultConfig.Authn.Logout.Query = `nav a[href="/signin/"]`
	defaultConfig.Authn.Logout.Label = "Logout"
	defaultConfig.Authn.Logout.Path = "/~/oauth/logout"
	defaultConfig.Importmap.PreloadModules = []string{
		"@kdex-ui",
	}

	transformer := &transform.AggregatedTransformer{
		Transformers: []transform.Transformer{
			&importmap.ImportMapTransformer{
				Config: &defaultConfig,
				ModuleImports: map[string]string{
					"@kdex-ui": "@kdex-ui/index.js",
				},
			},
			&app.AppTransformer{
				Config: &defaultConfig,
			},
			&meta.MetaTransformer{
				Config: &defaultConfig,
			},
			&navigation.NavigationTransformer{
				Config: &defaultConfig,
			},
		},
	}

	defaultConfig.Proxy.PathSeparator = "/_/"

	// Configure server for proxy
	s := Proxy{
		Config:      &defaultConfig,
		transformer: transformer,
	}

	// Create test proxy server
	proxyServer := httptest.NewServer(http.HandlerFunc(s.ReverseProxy()))
	defer proxyServer.Close()

	tests := []struct {
		name            string
		method          string
		path            string
		headers         map[string]string
		expectedStatus  int
		expectedBody    string
		upstreamAddress string
	}{
		{
			name:   "Basic GET request",
			method: "GET",
			path:   "/test/path",
			headers: map[string]string{
				"Authorization":     "Bearer test-token",
				"User-Agent":        "test-agent",
				"X-Forwarded-Host":  "foo.bar",
				"X-Forwarded-Proto": "http",
			},
			expectedStatus:  http.StatusOK,
			upstreamAddress: upstreamAddress,
		},
		{
			name:   "Basic GET request with query params",
			method: "GET",
			path:   "/test/path?param1=value1&param2=value2",
			headers: map[string]string{
				"Authorization":     "Bearer test-token",
				"User-Agent":        "test-agent",
				"X-Forwarded-Host":  "foo.bar",
				"X-Forwarded-Proto": "http",
			},
			expectedStatus:  http.StatusOK,
			upstreamAddress: upstreamAddress,
		},
		{
			name:            "Basic GET request returns 500",
			method:          "GET",
			path:            "/test/500",
			expectedStatus:  http.StatusInternalServerError,
			upstreamAddress: upstreamAddress,
		},
		{
			name:   "POST request",
			method: "POST",
			path:   "/api/data",
			headers: map[string]string{
				"Authorization":     "Bearer test-token-2",
				"User-Agent":        "test-agent-2",
				"X-Forwarded-Host":  "foo.bar",
				"X-Forwarded-Proto": "http",
			},
			expectedStatus:  http.StatusOK,
			upstreamAddress: upstreamAddress,
		},
		{
			name:            "GET request html without importmap",
			method:          "GET",
			path:            "/test/html_without_importmap",
			expectedStatus:  http.StatusOK,
			expectedBody:    `<html><head><script type="importmap">{"imports":{"@kdex-ui":"/~/m/@kdex-ui/index.js"}}</script><meta name="kdex-ui" data-path-separator="/_/" data-login-path="/~/oauth/login" data-login-label="Login" data-login-css-query="nav a[href=&#34;/signin/&#34;]" data-logout-path="/~/oauth/logout" data-logout-label="Logout" data-logout-css-query="nav a[href=&#34;/signin/&#34;]" data-state-endpoint="/~/state"/></head><body><h1>Hello, World!</h1><script type="module">import '@kdex-ui';</script></body></html>`,
			upstreamAddress: upstreamAddress,
		},
		{
			name:            "GET request html with importmap",
			method:          "GET",
			path:            "/test/html_with_importmap",
			expectedStatus:  http.StatusOK,
			expectedBody:    `<html><head><script type="importmap">{"imports":{"@foo/bar":"/foo/bar.js","@kdex-ui":"/~/m/@kdex-ui/index.js"}}</script><meta name="kdex-ui" data-path-separator="/_/" data-login-path="/~/oauth/login" data-login-label="Login" data-login-css-query="nav a[href=&#34;/signin/&#34;]" data-logout-path="/~/oauth/logout" data-logout-label="Logout" data-logout-css-query="nav a[href=&#34;/signin/&#34;]" data-state-endpoint="/~/state"/></head><body>test<script type="module">import '@kdex-ui';</script></body></html>`,
			upstreamAddress: upstreamAddress,
		},
		{
			name:            "GET with bad upstream address",
			method:          "GET",
			path:            "/test/html_with_importmap",
			expectedStatus:  http.StatusInternalServerError,
			expectedBody:    `Get "http://upstreamAddress/test/html_with_importmap": dial tcp: lookup upstreamAddress: Temporary failure in name resolution`,
			upstreamAddress: "upstreamAddress",
		},
		{
			name:            "GET path separator",
			method:          "GET",
			path:            "/test/_/foo/bar",
			expectedStatus:  http.StatusOK,
			expectedBody:    fmt.Sprintf(`{"method":"GET","path":"/test","headers":{"Accept-Encoding":["gzip"],"Forwarded":["by=%s;for=127.0.0.1;host=foo.bar;proto=http"],"User-Agent":["Go-http-client/1.1"],"X-Forwarded-For":["127.0.0.1"],"X-Forwarded-Host":["foo.bar"],"X-Forwarded-Proto":["http"],"X-Kdex-Proxy-App-Alias":["foo"],"X-Kdex-Proxy-App-Path":["/bar"]}}`, getOutboundIP().String()),
			upstreamAddress: upstreamAddress,
		},
		{
			name:            "GET path separator with html, route path matching app alias",
			method:          "GET",
			path:            "/test/app1/_/app1/bar",
			expectedStatus:  http.StatusOK,
			expectedBody:    `<html><head><script type="importmap">{"imports":{"@kdex-ui":"/~/m/@kdex-ui/index.js"}}</script><meta name="kdex-ui" data-path-separator="/_/" data-login-path="/~/oauth/login" data-login-label="Login" data-login-css-query="nav a[href=&#34;/signin/&#34;]" data-logout-path="/~/oauth/logout" data-logout-label="Logout" data-logout-css-query="nav a[href=&#34;/signin/&#34;]" data-state-endpoint="/~/state"/></head><body><kdex-ui-app-container><app-one id="app1" route-path="/bar"></app-one></kdex-ui-app-container><script type="module">import '@kdex-ui';</script><script type="module" src="http://localhost:61345/app1.js"></script></body></html>`,
			upstreamAddress: upstreamAddress,
		},
		{
			name:            "GET path separator with html, route path matching app alias but no app path",
			method:          "GET",
			path:            "/test/app1/_/app1",
			expectedStatus:  http.StatusOK,
			expectedBody:    `<html><head><script type="importmap">{"imports":{"@kdex-ui":"/~/m/@kdex-ui/index.js"}}</script><meta name="kdex-ui" data-path-separator="/_/" data-login-path="/~/oauth/login" data-login-label="Login" data-login-css-query="nav a[href=&#34;/signin/&#34;]" data-logout-path="/~/oauth/logout" data-logout-label="Logout" data-logout-css-query="nav a[href=&#34;/signin/&#34;]" data-state-endpoint="/~/state"/></head><body><kdex-ui-app-container><app-one id="app1"></app-one></kdex-ui-app-container><script type="module">import '@kdex-ui';</script><script type="module" src="http://localhost:61345/app1.js"></script></body></html>`,
			upstreamAddress: upstreamAddress,
		},
		{
			name:            "GET path separator with html, route path not matching app alias",
			method:          "GET",
			path:            "/test/app1/_/app2/bar",
			expectedStatus:  http.StatusOK,
			expectedBody:    `<html><head><script type="importmap">{"imports":{"@kdex-ui":"/~/m/@kdex-ui/index.js"}}</script><meta name="kdex-ui" data-path-separator="/_/" data-login-path="/~/oauth/login" data-login-label="Login" data-login-css-query="nav a[href=&#34;/signin/&#34;]" data-logout-path="/~/oauth/logout" data-logout-label="Logout" data-logout-css-query="nav a[href=&#34;/signin/&#34;]" data-state-endpoint="/~/state"/></head><body><kdex-ui-app-container><app-one id="app1"></app-one></kdex-ui-app-container><script type="module">import '@kdex-ui';</script><script type="module" src="http://localhost:61345/app1.js"></script></body></html>`,
			upstreamAddress: upstreamAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.Config.Proxy.UpstreamAddress = tt.upstreamAddress

			// Create request to proxy
			req, _ := http.NewRequest(tt.method, proxyServer.URL+tt.path, nil)

			// Add headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			req.Host = "foo.bar"
			req.URL.Scheme = "http"
			req.Header.Set("Host", req.Host)

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
				return
			}

			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				return
			}

			// Read and verify response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
				return
			}

			result := Result{}
			json.Unmarshal(body, &result)

			// Verify headers were properly forwarded
			for k, v := range tt.headers {
				if strings.Join(result.Headers[k], ",") != v {
					t.Errorf("Response doesn't contain expected header value %s: %s", k, v)
				}
			}

			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, resp.Header.Get("X-Test-Header"), "test-value")

				if tt.expectedBody != "" {
					assert.Equal(t, tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestServer_Probe(t *testing.T) {
	targetServer := mockTargetServer()
	defer targetServer.Close()

	upstreamAddress := strings.TrimPrefix(targetServer.URL, "http://")
	defaultConfig := config.DefaultConfig()
	defaultConfig.Proxy.UpstreamAddress = upstreamAddress
	defaultConfig.Proxy.UpstreamScheme = "http"
	defaultConfig.Proxy.UpstreamHealthzPath = "/healthz"

	s := &Proxy{
		Config: &defaultConfig,
	}

	proxyServer := httptest.NewServer(http.HandlerFunc(s.ReverseProxy()))
	defer proxyServer.Close()

	tests := []struct {
		name            string
		upstreamAddress string
		recorder        *httptest.ResponseRecorder
		method          string
		healthzPath     string
		wantBody        string
		wantStatus      int
	}{
		{
			name:            "test",
			upstreamAddress: upstreamAddress,
			recorder:        httptest.NewRecorder(),
			method:          "GET",
			healthzPath:     "/healthz",
			wantBody:        "OK",
			wantStatus:      http.StatusOK,
		},
		{
			name:            "test failed probe",
			upstreamAddress: "upstreamAddress",
			recorder:        httptest.NewRecorder(),
			method:          "GET",
			healthzPath:     "/healthz",
			wantBody:        `Get "http://upstreamAddress/healthz": dial tcp: lookup upstreamAddress: Temporary failure in name resolution`,
			wantStatus:      http.StatusInternalServerError,
		},
		{
			name:            "test failed probe with 400",
			upstreamAddress: upstreamAddress,
			recorder:        httptest.NewRecorder(),
			method:          "GET",
			healthzPath:     "/test/400",
			wantBody:        `GET /test/400 returned 400`,
			wantStatus:      http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.Config.Proxy.UpstreamAddress = tt.upstreamAddress
			s.Config.Proxy.UpstreamHealthzPath = tt.healthzPath

			s.Probe(tt.recorder, httptest.NewRequest(tt.method, proxyServer.URL, nil))

			assert.Equal(t, tt.recorder.Code, tt.wantStatus)
			assert.Equal(t, tt.recorder.Body.String(), tt.wantBody)
		})
	}
}

func TestServer_rewrite(t *testing.T) {
	defaultConfig := config.DefaultConfig()
	defaultConfig.Proxy.UpstreamAddress = "target-server"
	defaultConfig.Proxy.UpstreamScheme = "http"
	defaultConfig.Proxy.PathSeparator = "/_/"
	defaultConfig.Proxy.AlwaysAppendSlash = true
	defaultConfig.Navigation.TemplatePaths = []config.TemplatePath{
		{
			Href:     "/alias",
			Label:    "Alias",
			Template: "/template",
			Weight:   1,
		},
	}

	tests := []struct {
		name string
		r    *httputil.ProxyRequest
		want *url.URL
	}{
		{
			name: "test",
			r: &httputil.ProxyRequest{
				In: &http.Request{
					URL:    &url.URL{Path: "/test", Scheme: "http", Host: "localhost"},
					Header: http.Header{},
				},
				Out: &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				},
			},
			want: &url.URL{Path: "/test", Scheme: "http", Host: "target-server"},
		},
		{
			name: "test with alias",
			r: &httputil.ProxyRequest{
				In: &http.Request{
					URL:    &url.URL{Path: "/alias", Scheme: "http", Host: "localhost"},
					Header: http.Header{},
				},
				Out: &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				},
			},
			want: &url.URL{Path: "/template", Scheme: "http", Host: "target-server"},
		},
		{
			name: "test with alias and path",
			r: &httputil.ProxyRequest{
				In: &http.Request{
					URL:    &url.URL{Path: "/css/base.min.2fbd9dd903cac0d10e1ae4765ed55e6f79bf4e4728d27a56f74dae99768ca735.css", Scheme: "http", Host: "localhost"},
					Header: http.Header{},
				},
				Out: &http.Request{
					URL:    &url.URL{},
					Header: http.Header{},
				},
			},
			want: &url.URL{Path: "/css/base.min.2fbd9dd903cac0d10e1ae4765ed55e6f79bf4e4728d27a56f74dae99768ca735.css", Scheme: "http", Host: "target-server"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewProxy(
				&defaultConfig,
				nil,
			)
			s.rewrite(tt.r)
			assert.Equal(t, tt.want, tt.r.Out.URL)
		})
	}
}
