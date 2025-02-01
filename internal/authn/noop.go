package authn

import (
	"net/http"
)

type NoOpAuthValidator struct{}

func (v *NoOpAuthValidator) GetAuthType() string {
	return ""
}

func (v *NoOpAuthValidator) GetRealm() string {
	return ""
}

func (v *NoOpAuthValidator) Validate(r *http.Request) *AuthChallenge {
	return nil
}
