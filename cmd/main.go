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

func main() {
	ps := proxy.NewServerFromEnv()
	transformer := importmap.NewImportMapTransformerFromEnv()
	fs, err := fileserver.NewFileServerFromEnv()

	if err != nil {
		log.Fatal(err)
	}

	transformer.WithModulePrefix(fs.Prefix)
	ps.WithTransformer(transformer)

	mux := http.NewServeMux()

	mux.Handle("GET "+fs.Prefix, middlewareLogger(fs.ServeHTTP()))
	mux.Handle("GET "+ps.ProbePrefix, middlewareLogger(http.HandlerFunc(ps.Probe)))
	mux.Handle("/", middlewareLogger(http.HandlerFunc(ps.ReverseProxy())))

	log.Printf("Listening on %s:%s", ps.ListenAddress, ps.ListenPort)

	if err := http.ListenAndServe(ps.ListenAddress+":"+ps.ListenPort, mux); err != nil {
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
