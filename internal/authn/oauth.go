package authn

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"kdex.dev/proxy/internal/session"
	"kdex.dev/proxy/internal/util"
)

type OAuthValidator struct {
	AuthorizationHeader string
	ClientID            string
	ClientSecret        string
	Oauth2Config        *oauth2.Config
	Prefix              string
	Provider            *oidc.Provider
	Realm               string
	RedirectURI         string
	Scopes              []string
	SessionStore        session.SessionStore
	Verifier            *oidc.IDTokenVerifier
}

type Config struct {
	AuthorizationHeader string
	AuthServerURL       string
	ClientID            string
	ClientSecret        string
	Prefix              string
	Realm               string
	RedirectURI         string
	Scopes              []string
}

func NewOAuthValidator(ctx context.Context, config *Config) *OAuthValidator {
	providerURL := fmt.Sprintf("%s/realms/%s", config.AuthServerURL, url.PathEscape(config.Realm))
	log.Printf("Provider URL: %s", providerURL)
	provider, err := oidc.NewProvider(ctx, providerURL)

	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	config.Scopes = append(config.Scopes, oidc.ScopeOpenID)
	config.Scopes = append(config.Scopes, "roles")

	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.ClientID,
	})

	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       config.Scopes,
	}

	return &OAuthValidator{
		AuthorizationHeader: config.AuthorizationHeader,
		ClientID:            config.ClientID,
		ClientSecret:        config.ClientSecret,
		Oauth2Config:        &oauth2Config,
		Prefix:              config.Prefix,
		Provider:            provider,
		Realm:               config.Realm,
		RedirectURI:         config.RedirectURI,
		Scopes:              config.Scopes,
		SessionStore:        session.NewMemorySessionStore(),
		Verifier:            verifier,
	}
}

func (v *OAuthValidator) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+v.Prefix+"oauth/callback", v.callbackHandler())
	mux.HandleFunc("GET "+v.Prefix+"oauth/signin", v.signInHandler())
}

func (v *OAuthValidator) Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		return func(h http.Handler) {
			v.signInHandler()(w, r)
		}
	}

	sessionData, err := v.SessionStore.Get(r.Context(), sessionCookie.Value)
	if err != nil {
		return func(h http.Handler) {
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: "",
				Path:  "/",
			})
			v.signInHandler()(w, r)
		}
	}

	token, err := v.Provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true, // ?
	}).Verify(r.Context(), sessionData.AccessToken)

	if err != nil {
		return func(h http.Handler) {
			v.SessionStore.Delete(r.Context(), sessionCookie.Value)
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: "",
				Path:  "/",
			})
			v.signInHandler()(w, r)
		}
	}

	var claims map[string]interface{}
	if err := token.Claims(&claims); err != nil {
		return func(h http.Handler) {
			v.signInHandler()(w, r)
		}
	}

	return func(h http.Handler) {
		r = r.WithContext(context.WithValue(r.Context(), ContextUserKey, UserData{
			Claims:  claims,
			Session: sessionData,
		}))

		h.ServeHTTP(w, r)
	}
}

type UserData struct {
	Claims  map[string]interface{}
	Session *session.SessionData
}

func (v *OAuthValidator) callbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := v.verifyState(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		authorizationCode := r.URL.Query().Get("code")
		if authorizationCode == "" {
			http.Error(w, "authorizationCode is required", http.StatusBadRequest)
			return
		}

		opts := []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		}

		oauth2Token, err := v.Oauth2Config.Exchange(r.Context(), authorizationCode, opts...)
		if err != nil {
			log.Printf("Error exchanging authorization code: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo, err := v.validateAndGetClaimsIDToken(r.Context(), oauth2Token)
		if err != nil {
			log.Printf("Error validating and getting claims ID token: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("User info: %+v", userInfo)

		sessionID := util.RandStringBytes(32)

		sessionData := session.SessionData{
			AccessToken: oauth2Token.AccessToken,
			UserInfo: session.UserInfo{
				Username: userInfo.Username,
				Email:    userInfo.Email,
			},
			CreatedAt: time.Now(),
		}

		if err := v.SessionStore.Set(r.Context(), sessionID, sessionData); err != nil {
			log.Printf("Error setting session: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Name:     "session_id",
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Secure:   util.GetScheme(r) == "https",
			Value:    sessionID,
		})

		redirectURI := r.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			redirectURI = "/"
		}

		http.Redirect(w, r, redirectURI, http.StatusTemporaryRedirect)
	}
}

type oidcClaims struct {
	Email    string `json:"email"`
	Username string `json:"preferred_username"`
}

func (v *OAuthValidator) validateAndGetClaimsIDToken(ctx context.Context, oauth2Token *oauth2.Token) (*oidcClaims, error) {
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token is required")
	}
	idToken, err := v.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify id_token: %w", err)
	}
	claims := oidcClaims{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	return &claims, nil
}

func (v *OAuthValidator) verifyState(r *http.Request) error {
	state := r.URL.Query().Get("state")
	if state == "" {
		return fmt.Errorf("state is required")
	}

	// TODO add a more secure check of the state param
	// One option would be to use a signed token with a secret key
	if state != "test_state" {
		return fmt.Errorf("invalid state")
	}

	return nil
}

func (v *OAuthValidator) signInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := fmt.Sprintf("%s://%s%s", util.GetScheme(r), r.Host, v.Prefix+"oauth/callback")
		log.Printf("Redirect URL: %s", redirectURL)
		authURL := v.Oauth2Config.AuthCodeURL(
			"test_state",
			oauth2.SetAuthURLParam("redirect_uri", redirectURL),
		)
		log.Printf("Auth URL: %s", authURL)
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}
