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
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/importmap"
)

var server string
var port string
var mapped_headers []string

func main() {
	http.HandleFunc("/", handler)

	setup()

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func setup() {
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	server = os.Getenv("SERVER")
	if server == "" {
		log.Fatal("SERVER environment variable not set")
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

// handler processes requests.
func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		log.Printf("Request processed in %v", time.Since(start))
	}()

	// Default to http if scheme is empty
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	var url string
	if r.URL.RawQuery != "" {
		url = fmt.Sprintf("%s://%s%s?%s", scheme, server, r.URL.Path, r.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s://%s%s", scheme, server, r.URL.Path)
	}

	var reqBody io.Reader
	if r.Body != nil {
		reqBody = r.Body
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, _ := http.NewRequest(r.Method, url, reqBody)

	for _, header := range mapped_headers {
		mapHeader(r, req, header)
	}

	resp, err := client.Do(req)

	if err != nil {
		// this is a network error like a timeout, so we should return 500
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
