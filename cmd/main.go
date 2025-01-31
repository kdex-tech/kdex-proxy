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
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kdex.dev/proxy/internal/app"
	"kdex.dev/proxy/internal/fileserver"
	"kdex.dev/proxy/internal/importmap"
	"kdex.dev/proxy/internal/proxy"
	"kdex.dev/proxy/internal/transform"
)

func main() {
	ps := proxy.NewServerFromEnv()

	httpServer := &http.Server{Addr: ps.ListenAddress + ":" + ps.ListenPort}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
		log.Println("Server graceful shutdown complete.")
	}()

	fs, err := fileserver.NewFileServerFromEnv()

	if err != nil {
		log.Fatal(err)
	}

	transformer := &transform.AggregatedTransformer{
		Transformers: []transform.Transformer{
			importmap.NewImportMapTransformerFromEnv().WithModulePrefix(fs.Prefix),
			&app.AppTransformer{AppManager: app.NewAppManagerFromEnv()},
		},
	}

	ps.WithTransformer(transformer)

	mux := http.NewServeMux()

	mux.Handle("GET "+fs.Prefix, middlewareLogger(fs.ServeHTTP()))
	mux.Handle("GET "+ps.ProbePrefix, middlewareLogger(http.HandlerFunc(ps.Probe)))
	mux.Handle("/", middlewareLogger(http.HandlerFunc(ps.ReverseProxy())))

	httpServer.Handler = mux

	log.Printf("Server listening on %s:%s", ps.ListenAddress, ps.ListenPort)

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped gracefully.")
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
