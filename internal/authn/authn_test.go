package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/store/session"
)

func TestAuthValidatorFactory(t *testing.T) {
	type args struct {
		c                 *config.AuthnConfig
		sessionStore      *session.SessionStore
		sessionCookieName string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "constructor with default config",
			args: args{
				c: &config.AuthnConfig{},
			},
			want: &NoOpAuthValidator{},
		},
		{
			name: "constructor with static basic auth config",
			args: args{
				c: &config.AuthnConfig{AuthValidator: Validator_StaticBasicAuth},
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
			got := AuthValidatorFactory(tt.args.c, tt.args.sessionStore, tt.args.sessionCookieName)
			assert.Equal(t, tt.want, got)
		})
	}
}
