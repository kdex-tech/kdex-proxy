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
	"time"

	"kdex.dev/proxy/internal/transform"
	"kdex.dev/proxy/internal/util"
)

const (
	DefaultListenAddress       = ""
	DefaultListenPort          = "8080"
	DefaultProbePrefix         = "/~/p/{$}"
	DefaultUpstreamScheme      = "http"
	DefaultUpstreamHealthzPath = "/"
)

type Server struct {
	proxy               *httputil.ReverseProxy
	transformer         transform.Transformer
	ListenAddress       string
	ListenPort          string
	ProbePrefix         string
	UpstreamAddress     string
	UpstreamScheme      string
	UpstreamHealthzPath string
}

func NewServerFromEnv() *Server {
	listen_port := os.Getenv("LISTEN_PORT")
	if listen_port == "" {
		listen_port = DefaultListenPort
		log.Printf("Defaulting listen_port to %s", listen_port)
	}

	listen_address := os.Getenv("LISTEN_ADDRESS")
	if listen_address == "" {
		listen_address = DefaultListenAddress
		log.Printf("Defaulting listen_address to any address on all interfaces")
	}

	probe_prefix := os.Getenv("PROBE_PREFIX")
	if probe_prefix == "" {
		probe_prefix = DefaultProbePrefix
		log.Printf("Defaulting probe_prefix to %s", probe_prefix)
	}

	upstream_address := os.Getenv("UPSTREAM_ADDRESS")
	if upstream_address == "" {
		log.Fatal("UPSTREAM_ADDRESS environment variable not set")
	}

	upstream_scheme := os.Getenv("UPSTREAM_SCHEME")
	if upstream_scheme == "" {
		upstream_scheme = DefaultUpstreamScheme
		log.Printf("Defaulting upstream_scheme to %s", upstream_scheme)
	}

	upstream_healthz_path := os.Getenv("UPSTREAM_HEALTHZ_PATH")
	if upstream_healthz_path == "" {
		upstream_healthz_path = DefaultUpstreamHealthzPath
		log.Printf("Defaulting upstream_healthz_path to %s", upstream_healthz_path)
	}

	return &Server{
		ListenAddress:       listen_address,
		ListenPort:          listen_port,
		ProbePrefix:         probe_prefix,
		UpstreamAddress:     upstream_address,
		UpstreamScheme:      upstream_scheme,
		UpstreamHealthzPath: upstream_healthz_path,
	}
}

func (s *Server) WithTransformer(transformer transform.Transformer) *Server {
	s.transformer = transformer
	return s
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
	r.SetURL(&url.URL{Scheme: s.UpstreamScheme, Host: s.UpstreamAddress})
	r.Out.Host = r.In.Host
	r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
	r.SetXForwarded()
	setXForwardedPort(r.In, r.Out)
	setForwarded(r.In, r.Out)
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

	if err := (s.transformer).Transform(&b); err != nil {
		return err
	}

	r.Body = io.NopCloser(bytes.NewReader(b))
	r.ContentLength = int64(len(b))
	r.Header.Set("Content-Length", strconv.Itoa(len(b)))

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
