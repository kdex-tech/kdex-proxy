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
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/importmap"
	"kdex.dev/proxy/internal/util"
)

type Server struct {
	ListenAddress       string
	ListenPort          string
	UpstreamAddress     string
	UpstreamScheme      string
	UpstreamHealthzPath string
}

func NewServerFromEnv() *Server {
	listen_port := os.Getenv("LISTEN_PORT")
	if listen_port == "" {
		listen_port = "8080"
		log.Printf("Defaulting listen_port to %s", listen_port)
	}

	listen_address := os.Getenv("LISTEN_ADDRESS")
	if listen_address == "" {
		log.Print("Defaulting listen_address to none (any address)")
	}

	upstream_address := os.Getenv("UPSTREAM_ADDRESS")
	if upstream_address == "" {
		log.Fatal("UPSTREAM_ADDRESS environment variable not set")
	}

	upstream_scheme := os.Getenv("UPSTREAM_SCHEME")
	if upstream_scheme == "" {
		upstream_scheme = "http"
	}

	upstream_healthz_path := os.Getenv("UPSTREAM_HEALTHZ_PATH")
	if upstream_healthz_path == "" {
		upstream_healthz_path = "/"
	}

	return &Server{
		ListenAddress:       listen_address,
		ListenPort:          listen_port,
		UpstreamAddress:     upstream_address,
		UpstreamScheme:      upstream_scheme,
		UpstreamHealthzPath: upstream_healthz_path,
	}
}

func (s *Server) ReverseProxy() func(http.ResponseWriter, *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(
		&url.URL{Scheme: s.UpstreamScheme, Host: s.UpstreamAddress})

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	proxy.Director = func(req *http.Request) {
		processProxyHeaders(req, req)
	}

	proxy.ModifyResponse = modifyResponse

	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			r.URL.Host = s.UpstreamAddress
			r.URL.Scheme = s.UpstreamScheme
			p.ServeHTTP(w, r)
		}
	}

	return handler(proxy)
}

func modifyResponse(r *http.Response) (err error) {
	// Check if response is HTML and not streaming
	contentType := r.Header.Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isStreaming := r.Header.Get("Transfer-Encoding") == "chunked"

	if !isHTML || isStreaming {
		return nil
	}

	// Buffer HTML content
	b, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	if err = r.Body.Close(); err != nil {
		return err
	}

	if err := mutateImportMap(&b); err != nil {
		return err
	}

	r.Body = io.NopCloser(bytes.NewReader(b))
	r.ContentLength = int64(len(b))
	r.Header.Set("Content-Length", strconv.Itoa(len(b)))

	return nil
}

func mutateImportMap(body *[]byte) error {
	doc, err := html.Parse(bytes.NewReader(*body))
	if err != nil {
		return err
	}

	importMapManager := importmap.Manager(doc).WithMutator(
		func(importMap *importmap.ImportMap) {
			importMap.Imports["@kdex-ui"] = "/_/kdex-ui.js"
		},
	)

	if !importMapManager.Mutate() {
		return nil
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		log.Printf("Error rendering modified HTML: %v", err)
		return err
	}
	*body = buf.Bytes()
	return nil
}

func processProxyHeaders(r *http.Request, req *http.Request) {
	setXForwardedFor(r, req)

	if strings.Contains(r.Host, ":") {
		hostName, _, _ := net.SplitHostPort(r.Host)
		req.Header.Set("X-Forwarded-Host", hostName)
	} else {
		req.Header.Set("X-Forwarded-Host", r.Host)
	}

	req.Header.Set("X-Forwarded-Proto", util.GetScheme(r))

	if strings.Contains(r.Host, ":") {
		if _, port, _ := net.SplitHostPort(r.Host); port != "" {
			req.Header.Set("X-Forwarded-Port", port)
		}
	}

	setForwarded(r, req)
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func setForwarded(r *http.Request, req *http.Request) {
	req.Header.Set("Forwarded", fmt.Sprintf("by=%s;for=%s;host=%s;proto=%s", getOutboundIP().String(), r.RemoteAddr, r.Host, util.GetScheme(r)))
}

func setXForwardedFor(r *http.Request, req *http.Request) {
	if xForwardedFor := r.Header.Values("X-Forwarded-For"); len(xForwardedFor) > 0 {
		xForwardedFors := strings.Join(xForwardedFor, ", ")
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("%s, %s", xForwardedFors, getOutboundIP().String()))
	} else {
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("%s, %s", r.RemoteAddr, getOutboundIP().String()))
	}
}
