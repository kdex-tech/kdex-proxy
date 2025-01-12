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
	"log"
	"net/http"
	"time"

	"kdex.dev/proxy/internal/fileserver"
	"kdex.dev/proxy/internal/importmap"
	"kdex.dev/proxy/internal/proxy"
)

var s *proxy.Server

func main() {
	s = proxy.NewServerFromEnv()

	s.WithTransformer(importmap.NewImportMapTransformer())

	fs, err := fileserver.NewFileServerFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.Handle("GET "+fs.Prefix, middlewareLogger(fs.ServeHTTP()))
	mux.Handle("GET "+s.ProbePrefix, http.HandlerFunc(s.Probe))
	mux.Handle("/", middlewareLogger(http.HandlerFunc(s.ReverseProxy())))

	log.Printf("Listening on %s:%s", s.ListenAddress, s.ListenPort)

	if err := http.ListenAndServe(s.ListenAddress+":"+s.ListenPort, mux); err != nil {
		log.Fatal(err)
	}
}

func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.Printf("%s %s processed in %v", r.Method, r.URL.Path, time.Since(start))
		}()

		next.ServeHTTP(w, r)
	})
}
