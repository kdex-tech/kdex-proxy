package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// mockTargetServer creates a test server that simulates the target server
func mockTargetServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Test-Header", "test-value")

		if r.URL.Path == "/test/400" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if r.URL.Path == "/test/500" {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `{
			"method": "%s",
			"path": "%s",
			"headers": {
				"Authorization": "%s",
				"User-Agent": "%s"
			}
		}`, r.Method, r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("User-Agent"))
	}))
}

func TestProxyEndToEnd(t *testing.T) {
	// Start mock target server
	targetServer := mockTargetServer()
	defer targetServer.Close()

	// Configure environment for proxy
	os.Setenv("SERVER", strings.TrimPrefix(targetServer.URL, "http://"))
	os.Setenv("MAPPED_HEADERS", "Authorization,User-Agent")
	setup()

	// Create test proxy server
	proxyServer := httptest.NewServer(http.HandlerFunc(handler))
	defer proxyServer.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:   "Basic GET request",
			method: "GET",
			path:   "/test/path",
			headers: map[string]string{
				"Authorization": "Bearer test-token",
				"User-Agent":    "test-agent",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Basic GET request with query params",
			method: "GET",
			path:   "/test/path?param1=value1&param2=value2",
			headers: map[string]string{
				"Authorization": "Bearer test-token",
				"User-Agent":    "test-agent",
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
				"Authorization": "Bearer test-token-2",
				"User-Agent":    "test-agent-2",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request to proxy
			req, err := http.NewRequest(tt.method, proxyServer.URL+tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Read and verify response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			// Verify headers were properly forwarded
			for k, v := range tt.headers {
				if !strings.Contains(string(body), v) {
					t.Errorf("Response doesn't contain expected header value %s: %s", k, v)
				}
			}

			if resp.StatusCode == http.StatusOK {
				// Verify response headers from target were forwarded
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Error("Content-Type header not properly forwarded")
				}
				if resp.Header.Get("X-Test-Header") != "test-value" {
					t.Error("X-Test-Header not properly forwarded")
				}
			}
		})
	}
}
