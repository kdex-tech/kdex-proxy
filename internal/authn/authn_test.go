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
