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
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/transform"
	"kdex.dev/proxy/internal/util"
)

type Server struct {
	proxy               *httputil.ReverseProxy
	transformer         transform.Transformer
	AlwaysAppendSlash   bool
	PathSeparator       string
	ProbePath           string
	UpstreamAddress     string
	UpstreamScheme      string
	UpstreamHealthzPath string
}

func NewServer(proxyConfig *config.ProxyConfig, transformer transform.Transformer) *Server {
	return &Server{
		AlwaysAppendSlash:   proxyConfig.AlwaysAppendSlash,
		PathSeparator:       proxyConfig.PathSeparator,
		ProbePath:           proxyConfig.ProbePath,
		UpstreamAddress:     proxyConfig.UpstreamAddress,
		UpstreamScheme:      proxyConfig.UpstreamScheme,
		UpstreamHealthzPath: proxyConfig.UpstreamHealthzPath,
		transformer:         transformer,
	}
}

func (s *Server) Probe(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s://%s%s", util.GetScheme(r), s.UpstreamAddress, s.UpstreamHealthzPath)

	req, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		w.Write([]byte("OK"))
	} else {
		w.Write([]byte(fmt.Sprintf("GET %s returned %d", s.UpstreamHealthzPath, resp.StatusCode)))
	}
}

func (s *Server) ReverseProxy() func(http.ResponseWriter, *http.Request) {
	s.proxy = &httputil.ReverseProxy{
		ErrorHandler:   s.errorHandler,
		ModifyResponse: s.modifyResponse,
		Rewrite:        s.rewrite,
	}

	return s.proxy.ServeHTTP
}

func (s *Server) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Error: %v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Server) rewrite(r *httputil.ProxyRequest) {
	target := &url.URL{Scheme: s.UpstreamScheme, Host: s.UpstreamAddress}
	req := r.Out

	parts := s.rewritePath(r)

	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host

	if parts.AppAlias != "" {
		req.URL.Path = parts.Path
		log.Printf("Path rewritten as: %s to %s (Alias: %s, Path: %s)", r.In.URL.Path, parts.Path, parts.AppAlias, parts.AppPath)
	} else {
		req.URL.Path = r.In.URL.Path
	}

	req.URL.RawPath = r.In.URL.RawPath

	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}

	r.Out.Host = r.In.Host

	r.Out.Header["X-Kdex-Proxy-App-Alias"] = []string{parts.AppAlias}
	r.Out.Header["X-Kdex-Proxy-App-Path"] = []string{parts.AppPath}

	{
		// Everything bellow is about Proxy Protocol
		r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
		r.SetXForwarded()
		setXForwardedPort(r.In, r.Out)
		setForwarded(r.In, r.Out)
	}
}

type PathParts struct {
	AppAlias string
	AppPath  string
	Path     string
}

func (s *Server) rewritePath(r *httputil.ProxyRequest) PathParts {
	parts := strings.SplitN(r.In.URL.Path, s.PathSeparator, 2)

	if len(parts) > 1 {
		newPath := parts[0]

		if s.AlwaysAppendSlash && !strings.HasSuffix(newPath, "/") {
			newPath = newPath + "/"
		}

		appParts := strings.SplitN(parts[1], "/", 2)

		if len(appParts) > 1 {
			appAlias := appParts[0]
			appPath := fmt.Sprintf("/%s", appParts[1])
			return PathParts{
				AppAlias: appAlias,
				AppPath:  appPath,
				Path:     newPath,
			}
		} else {
			return PathParts{
				AppAlias: parts[1],
				AppPath:  "",
				Path:     newPath,
			}
		}
	}

	return PathParts{
		AppAlias: "",
		AppPath:  "",
		Path:     r.In.URL.Path,
	}
}

func setXForwardedPort(in *http.Request, out *http.Request) {
	if strings.Contains(in.Host, ":") {
		_, port, err := net.SplitHostPort(in.Host)
		if err != nil {
			log.Printf("Error splitting host and port from %s: %v", in.Host, err)
			return
		}
		out.Header.Set("X-Forwarded-Port", port)
	}
}

func (s *Server) modifyResponse(r *http.Response) (err error) {
	if s.transformer == nil {
		return nil
	}

	if !(s.transformer).ShouldTransform(r) {
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

	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return err
	}

	if err := (s.transformer).Transform(r, doc); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		log.Printf("Error rendering modified HTML: %v", err)
		return err
	}

	b = buf.Bytes()
	contentLength := len(b)
	r.Body = io.NopCloser(bytes.NewReader(b))
	r.ContentLength = int64(contentLength)
	r.Header.Set("Content-Length", strconv.Itoa(contentLength))

	return nil
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

func setForwarded(in *http.Request, out *http.Request) {
	out.Header.Set(
		"Forwarded",
		fmt.Sprintf(
			"by=%s;for=%s;host=%s;proto=%s",
			getOutboundIP().String(),
			out.Header.Get("X-Forwarded-For"),
			in.Host,
			util.GetScheme(in)))
}
