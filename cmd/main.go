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
	"fmt"
	"log"
	"net/http"
	"time"

	"kdex.dev/proxy/internal/proxy"
	"kdex.dev/proxy/internal/util"
)

var s *proxy.Server

func main() {
	s = proxy.NewServerFromEnv()

	mux := http.NewServeMux()

	mux.Handle("/.proxy.probe", http.HandlerFunc(probe))
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

func probe(w http.ResponseWriter, r *http.Request) {
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
