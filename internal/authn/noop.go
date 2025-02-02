package authn

import (
	"net/http"
)

type NoOpAuthValidator struct{}

func (v *NoOpAuthValidator) Validate(r *http.Request) *AuthChallenge {
	return nil
}
