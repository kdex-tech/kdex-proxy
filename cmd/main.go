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
	"kdex.dev/proxy/internal/authn"
	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/fileserver"
	"kdex.dev/proxy/internal/importmap"
	"kdex.dev/proxy/internal/meta"
	mAuthn "kdex.dev/proxy/internal/middleware/authn"
	mAuthz "kdex.dev/proxy/internal/middleware/authz"
	mLogger "kdex.dev/proxy/internal/middleware/log"
	mRoles "kdex.dev/proxy/internal/middleware/roles"
	"kdex.dev/proxy/internal/navigation"
	"kdex.dev/proxy/internal/proxy"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/transform"
)

func main() {
	c := config.NewConfigFromEnv()

	httpServer := &http.Server{
		Addr: c.ListenAddress + ":" + c.ListenPort,
	}
	fileServer := fileserver.FileServer{
		Dir:    c.ModuleDir,
		Prefix: c.Fileserver.Prefix,
	}
	sessionStore, err := session.NewSessionStore(context.Background(), &c.Session)
	if err != nil {
		log.Fatalf("Failed to create session store: %v", err)
	}

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

	sessionHelper := &session.SessionHelper{
		Config:       &c,
		SessionStore: &sessionStore,
	}

	transformer := &transform.AggregatedTransformer{
		Transformers: []transform.Transformer{
			importmap.NewImportMapTransformer(&c),
			meta.NewMetaTransformer(&c),
			navigation.NewNavigationTransformer(&c, sessionHelper),
			app.NewAppTransformer(&c),
		},
	}

	proxyServer := proxy.NewServer(&c, transformer)

	mux := http.NewServeMux()

	authValidator := authn.AuthValidatorFactory(
		&c.Authn,
		&sessionStore,
		c.Session.CookieName,
	)

	loggerMiddleware := &mLogger.LoggerMiddleware{
		Impl: log.Default(),
	}

	authValidator.Register(mux)

	authnMiddleware := &mAuthn.AuthnMiddleware{
		AuthenticateHeader:     c.Authn.AuthenticateHeader,
		AuthenticateStatusCode: c.Authn.AuthenticateStatusCode,
		ProtectedPaths:         c.Navigation.ProtectedPaths,
		AuthValidator:          authValidator,
	}

	fieldEvaluator := expression.NewFieldEvaluator(&c)

	// After authn middleware
	rolesMiddleware := &mRoles.RolesMiddleware{
		FieldEvaluator: fieldEvaluator,
	}

	// Create authorizer with provider
	permProvider := authz.NewPermissionProvider(&c)
	authorizer := authz.NewAuthorizer(permProvider)
	authzMiddleware := &mAuthz.AuthzMiddleware{
		Authorizer: authorizer,
	}

	stateHandler := &authn.StateHandler{
		FieldEvaluator: fieldEvaluator,
	}

	mux.Handle("GET "+c.Fileserver.Prefix, loggerMiddleware.Log(fileServer.ServeHTTP()))
	mux.Handle("GET "+c.Proxy.ProbePath, loggerMiddleware.Log(http.HandlerFunc(proxyServer.Probe)))
	mux.Handle("GET "+c.Authn.StateEndpoint,
		loggerMiddleware.Log(
			authnMiddleware.Authn(
				rolesMiddleware.InjectRoles(
					authzMiddleware.Authz(
						stateHandler.StateHandler()),
				),
			),
		),
	)
	mux.Handle("/",
		loggerMiddleware.Log(
			authnMiddleware.Authn(
				rolesMiddleware.InjectRoles(
					authzMiddleware.Authz(
						http.HandlerFunc(proxyServer.ReverseProxy()),
					),
				),
			),
		),
	)

	httpServer.Handler = mux

	log.Printf("Server listening on %s:%s", c.ListenAddress, c.ListenPort)

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped gracefully.")
}
