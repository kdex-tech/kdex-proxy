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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/proxy"
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

		if r.URL.Path == "/test/html_without_importmap" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body><h1>Hello, World!</h1></body></html>`))
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

func Test_reverseProxy(t *testing.T) {
	// Start mock target server
	targetServer := mockTargetServer()
	defer targetServer.Close()

	// Configure server for proxy
	s := proxy.Server{
		ListenAddress:       "localhost",
		ListenPort:          "8080",
		UpstreamAddress:     strings.TrimPrefix(targetServer.URL, "http://"),
		UpstreamScheme:      "http",
		UpstreamHealthzPath: "/healthz",
	}

	// Create test proxy server
	proxyServer := httptest.NewServer(http.HandlerFunc(s.ReverseProxy()))
	defer proxyServer.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
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
			expectedStatus: http.StatusOK,
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
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Basic GET request returns 500",
			method:         "GET",
			path:           "/test/500",
			expectedStatus: http.StatusInternalServerError,
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
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET request html without importmap",
			method:         "GET",
			path:           "/test/html_without_importmap",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body><h1>Hello, World!</h1></body></html>",
		},
		{
			name:           "GET request html with importmap",
			method:         "GET",
			path:           "/test/html_with_importmap",
			expectedStatus: http.StatusOK,
			expectedBody:   `<html><head><script type="importmap">{"imports":{"@foo/bar":"/foo/bar.js","@kdex-ui":"/_/kdex-ui.js"}}</script></head><body>test</body></html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				// Verify response headers from target were forwarded
				if resp.Header.Get("X-Test-Header") != "test-value" {
					t.Error("X-Test-Header not properly forwarded")
				}

				if tt.expectedBody != "" {
					assert.Equal(t, tt.expectedBody, string(body))
				}
			}
		})
	}
}
