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

package state

import (
	"context"
	"fmt"

	"kdex.dev/proxy/internal/config"
)

type StateStore interface {
	Set(ctx context.Context, state string) error
	Get(ctx context.Context, state string) (string, error)
	Delete(ctx context.Context, state string) error
}

func NewStateStore(config *config.Config) (StateStore, error) {
	switch config.State.Type {
	case "memory":
		return NewMemoryStateStore(config.State.TTL), nil
	}
	return nil, fmt.Errorf("invalid store type: %s", config.State.Type)
}
