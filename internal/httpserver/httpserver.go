// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kdex.dev/proxy/internal/config"
)

type HttpServer struct {
	server *http.Server
}

func NewHttpServer(config *config.Config) *HttpServer {
	server := &http.Server{
		Addr: config.ListenAddress + ":" + config.ListenPort,
	}

	return &HttpServer{
		server: server,
	}
}

func (s *HttpServer) SetHandler(mux *http.ServeMux) {
	s.server.Handler = mux
}

func (s *HttpServer) Start() error {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()

		log.Println("server graceful shutdown started.")

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}

		log.Println("server graceful shutdown complete.")
	}()

	log.Printf("server listening on %s", s.server.Addr)

	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %v", err)
	}

	log.Println("server stopped gracefully.")

	return nil
}

func (s *HttpServer) Stop() error {
	log.Println("server direct shutdown started.")

	if err := s.server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("server shutdown error: %v", err)
	}

	log.Println("server direct shutdown complete.")

	return nil
}
