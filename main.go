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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var server string
var mapped_headers []string

func main() {
	http.HandleFunc("/*", handler)

	port := os.Getenv("PORT")
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

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// handler processes requests.
func handler(w http.ResponseWriter, r *http.Request) {
	var url string
	if r.URL.RawQuery != "" {
		url = fmt.Sprintf("%s://%s%s?%s", r.URL.Scheme, server, r.URL.Path, r.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s://%s%s", r.URL.Scheme, server, r.URL.Path)
	}

	var reqBody io.Reader
	if r.Body != nil {
		reqBody = r.Body
	}

	client := &http.Client{}
	req, err := http.NewRequest(r.Method, url, reqBody)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, header := range mapped_headers {
		mapHeader(r, req, header)
	}

	resp, err := client.Do(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func mapHeader(r *http.Request, req *http.Request, headerName string) {
	headerValue := r.Header.Get(headerName)

	if headerValue != "" {
		req.Header.Set(headerName, headerValue)
	}
}
