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

package engine

import (
	"log"
	"net/http"

	"kdex.dev/proxy/internal/authn"
	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/check"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/fileserver"
	mAuthn "kdex.dev/proxy/internal/middleware/authn"
	mAuthz "kdex.dev/proxy/internal/middleware/authz"
	mLogger "kdex.dev/proxy/internal/middleware/log"
	mRoles "kdex.dev/proxy/internal/middleware/roles"
	"kdex.dev/proxy/internal/proxy"
	"kdex.dev/proxy/internal/state"

	"kdex.dev/proxy/internal/httpserver"
)

type Engine struct {
	Config     *config.Config
	httpServer *httpserver.HttpServer
}

func NewEngine(config *config.Config) *Engine {
	httpServer := httpserver.NewHttpServer(config)
	mux := http.NewServeMux()
	httpServer.SetHandler(mux)

	engine := &Engine{
		Config:     config,
		httpServer: httpServer,
	}

	// Components
	checker := check.NewChecker(config)
	authorizer := authz.NewAuthorizer(checker)
	authValidator := authn.AuthValidatorFactory(config)
	authValidator.Register(mux)
	fieldEvaluator := expression.NewFieldEvaluator(config)
	fileServer := fileserver.NewFileServer(config)
	proxyServer := proxy.NewProxy(config)
	stateHandler := &state.StateHandler{FieldEvaluator: fieldEvaluator}

	// Middleware
	authnMiddleware := &mAuthn.AuthnMiddleware{
		AuthenticateHeader:     config.Authn.AuthenticateHeader,
		AuthenticateStatusCode: config.Authn.AuthenticateStatusCode,
		AuthValidator:          authValidator,
	}
	loggerMiddleware := &mLogger.LoggerMiddleware{
		Impl: log.Default(),
	}
	rolesMiddleware := &mRoles.RolesMiddleware{
		FieldEvaluator: fieldEvaluator,
	}
	authzMiddleware := &mAuthz.AuthzMiddleware{
		Authorizer: authorizer,
	}

	// Handlers
	mux.Handle(
		"GET "+config.Fileserver.Prefix,
		loggerMiddleware.Log(fileServer.ServeHTTP(), false),
	)

	mux.Handle(
		"GET "+config.Proxy.ProbePath,
		loggerMiddleware.Log(http.HandlerFunc(proxyServer.Probe), true),
	)

	mux.Handle(
		"GET "+config.Authz.Endpoints.Single,
		loggerMiddleware.Log(
			authnMiddleware.Authn(
				rolesMiddleware.InjectRoles(
					authzMiddleware.Authz(
						checker.SingleHandler(),
					),
				),
			),
			false,
		),
	)

	mux.Handle(
		"POST "+config.Authz.Endpoints.Batch,
		loggerMiddleware.Log(
			authnMiddleware.Authn(
				rolesMiddleware.InjectRoles(
					authzMiddleware.Authz(
						checker.BatchHandler(),
					),
				),
			),
			false,
		),
	)

	mux.Handle(
		"GET "+config.State.Endpoint,
		loggerMiddleware.Log(
			authnMiddleware.Authn(
				rolesMiddleware.InjectRoles(
					authzMiddleware.Authz(
						stateHandler.StateHandler(),
					),
				),
			),
			false,
		),
	)

	mux.Handle(
		"/",
		loggerMiddleware.Log(
			authnMiddleware.Authn(
				rolesMiddleware.InjectRoles(
					authzMiddleware.Authz(
						http.HandlerFunc(proxyServer.ReverseProxy()),
					),
				),
			),
			false,
		),
	)

	return engine
}

func (e *Engine) Start() error {
	e.httpServer.Start()
	return nil
}

func (e *Engine) Stop() error {
	return nil
}
