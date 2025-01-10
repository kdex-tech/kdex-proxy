// Copyright 2023 Google LLC
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
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/importmap"
)

var listen_address string = ""
var listen_port string = "8080"
var mapped_headers []string
var upstream_address string
var upstream_healthz_path string = "/"

func main() {
	mux := http.NewServeMux()

	mainHandler := http.HandlerFunc(handler)
	probeHandler := http.HandlerFunc(probe)

	mux.Handle("/", middlewareLogger(mainHandler))
	mux.Handle("/.proxy.probe", probeHandler)

	setup()

	log.Printf("Listening on %s:%s", listen_address, listen_port)
	if err := http.ListenAndServe(listen_address+":"+listen_port, mux); err != nil {
		log.Fatal(err)
	}
}

func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			code := r.Header.Get("Response-Code")
			if code != "" {
				log.Printf("%s %s %s processed in %v", r.Method, r.URL.Path, code, time.Since(start))
			} else {
				log.Printf("%s %s processed in %v", r.Method, r.URL.Path, time.Since(start))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func setup() {
	listen_port = os.Getenv("LISTEN_PORT")
	if listen_port == "" {
		listen_port = "8080"
		log.Printf("Defaulting listen_port to %s", listen_port)
	}

	listen_address = os.Getenv("LISTEN_ADDRESS")
	if listen_address == "" {
		log.Print("Defaulting listen_address to none (any address)")
	}

	upstream_address = os.Getenv("UPSTREAM_ADDRESS")
	if upstream_address == "" {
		log.Fatal("UPSTREAM_ADDRESS environment variable not set")
	}

	upstream_healthz_path = os.Getenv("UPSTREAM_HEALTHZ_PATH")
	if upstream_healthz_path == "" {
		upstream_healthz_path = "/"
	}

	mapped_headers = strings.Split(os.Getenv("MAPPED_HEADERS"), ",")
	if len(mapped_headers) == 0 {
		mapped_headers = []string{
			"Authorization",
			"Content-Type",
			"Host",
			"Origin",
			"User-Agent",
		}
	}
}

func probe(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s://%s%s", getScheme(r), upstream_address, upstream_healthz_path)

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
		w.Write([]byte(fmt.Sprintf("GET %s returned %d", upstream_healthz_path, resp.StatusCode)))
	}
}

// handler processes requests.
func handler(w http.ResponseWriter, r *http.Request) {
	scheme := getScheme(r)

	var url string
	if r.URL.RawQuery != "" {
		url = fmt.Sprintf("%s://%s%s?%s", scheme, upstream_address, r.URL.Path, r.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s://%s%s", scheme, upstream_address, r.URL.Path)
	}

	var reqBody io.Reader
	if r.Body != nil {
		reqBody = r.Body
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, _ := http.NewRequest(r.Method, url, reqBody)

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add proxy headers to request
	processProxyHeaders(r, req)

	resp, err := client.Do(req)

	if err != nil {
		// This is a network error like a timeout, so we should return 500
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// Copy response headers first
	for key, values := range resp.Header {
		if key == "Content-Length" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code
	w.WriteHeader(resp.StatusCode)
	r.Header.Set("Response-Code", strconv.Itoa(resp.StatusCode))

	// Check if response is HTML and not streaming
	contentType := resp.Header.Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isStreaming := resp.Header.Get("Transfer-Encoding") == "chunked"

	if isHTML && !isStreaming {
		// Buffer HTML content
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading HTML body: %v", err)
			return
		}
		if err := mutateImportMap(&body); err != nil {
			log.Printf("Error mutating import map: %v", err)
			return
		}
		if _, err := w.Write(body); err != nil {
			log.Printf("Error writing HTML body: %v", err)
		}
	} else {
		// Stream other content types
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("Error copying response body: %v", err)
		}
	}
}

func processProxyHeaders(r *http.Request, req *http.Request) {
	setXForwardedFor(r, req)

	if strings.Contains(r.Host, ":") {
		hostName, _, _ := net.SplitHostPort(r.Host)
		req.Header.Set("X-Forwarded-Host", hostName)
	} else {
		req.Header.Set("X-Forwarded-Host", r.Host)
	}

	req.Header.Set("X-Forwarded-Proto", getScheme(r))

	if strings.Contains(r.Host, ":") {
		if _, port, _ := net.SplitHostPort(r.Host); port != "" {
			req.Header.Set("X-Forwarded-Port", port)
		}
	}

	setForwarded(r, req)
}

func mapHeader(r *http.Request, req *http.Request, headerName string) {
	headerValue := r.Header.Get(headerName)

	if headerValue != "" {
		req.Header.Set(headerName, headerValue)
	}
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

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func getScheme(r *http.Request) string {
	if r.URL.Scheme != "" {
		return r.URL.Scheme
	}
	return "http"
}

func setForwarded(r *http.Request, req *http.Request) {
	req.Header.Set("Forwarded", fmt.Sprintf("by=%s;for=%s;host=%s;proto=%s", getOutboundIP().String(), r.RemoteAddr, r.Host, getScheme(r)))
}

func setXForwardedFor(r *http.Request, req *http.Request) {
	if xForwardedFor := r.Header.Values("X-Forwarded-For"); len(xForwardedFor) > 0 {
		xForwardedFors := strings.Join(xForwardedFor, ", ")
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("%s, %s", xForwardedFors, getOutboundIP().String()))
	} else {
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("%s, %s", r.RemoteAddr, getOutboundIP().String()))
	}
}
