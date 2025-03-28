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

package authz

import (
	"errors"
	"fmt"
	"net/http"

	"kdex.dev/proxy/internal/check"
	"kdex.dev/proxy/internal/permission"
)

// Authorizer defines the interface for authorization checks
type Authorizer interface {
	// CheckAccess verifies if the request has permission to access the resource
	CheckAccess(r *http.Request) error
}

// NewAuthorizer creates a new Authorizer based on configuration
func NewAuthorizer(checker *check.Checker) Authorizer {
	return &defaultAuthorizer{
		checker: checker,
	}
}

type defaultAuthorizer struct {
	checker *check.Checker
}

func (a *defaultAuthorizer) CheckAccess(r *http.Request) error {
	ok, err := a.checker.Check(r.Context(), fmt.Sprintf("page:%s", r.URL.Path), "read")

	if err != nil && !errors.Is(err, permission.ErrNoPermissions) && !errors.Is(err, check.ErrNoRoles) {
		return err
	}

	if ok {
		return nil
	}

	return ErrUnauthorized
}
