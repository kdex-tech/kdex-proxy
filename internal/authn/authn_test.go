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

package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
)

func TestAuthValidatorFactory(t *testing.T) {
	tests := []struct {
		name string
		c    *config.Config
		want interface{}
	}{
		{
			name: "constructor with default config",
			c: &config.Config{
				Authn: config.AuthnConfig{
					AuthValidator: Validator_NoOp,
				},
			},
			want: &NoOpAuthValidator{},
		},
		{
			name: "constructor with static basic auth config",
			c: &config.Config{
				Authn: config.AuthnConfig{
					AuthValidator: Validator_StaticBasicAuth,
				},
			},
			want: &StaticBasicAuthValidator{},
		},
		// {
		// 	name: "constructor with oauth config",
		// 	args: args{
		// 		c: &config.AuthnConfig{AuthValidator: Validator_OAuth},
		// 	},
		// 	want: &OAuthValidator{},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AuthValidatorFactory(tt.c)
			assert.Equal(t, tt.want, got)
		})
	}
}
