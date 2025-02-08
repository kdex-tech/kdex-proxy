package authn

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/store/state"
	"kdex.dev/proxy/internal/util"
)

type Config struct {
	AuthorizationHeader string
	AuthServerURL       string
	ClientID            string
	ClientSecret        string
	DumpClaims          bool
	Prefix              string
	Realm               string
	RedirectURI         string
	Scopes              []string
	SignInOnChallenge   bool
}

type OAuthValidator struct {
	Config       *Config
	Oauth2Config *oauth2.Config
	Provider     *oidc.Provider
	SessionStore session.SessionStore
	StateStore   state.StateStore
	Verifier     *oidc.IDTokenVerifier
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

	stateStore, err := state.NewStateStore(ctx, "memory")
	if err != nil {
		log.Fatalf("Failed to create state store: %v", err)
	}

	sessionStore, err := session.NewSessionStore(ctx, "memory")
	if err != nil {
		log.Fatalf("Failed to create session store: %v", err)
	}

	return &OAuthValidator{
		Config:       config,
		Oauth2Config: &oauth2Config,
		Provider:     provider,
		SessionStore: sessionStore,
		StateStore:   stateStore,
		Verifier:     verifier,
	}
}

func (v *OAuthValidator) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+v.Config.Prefix+"oauth/callback", v.callbackHandler())
	mux.HandleFunc("GET "+v.Config.Prefix+"oauth/login", v.logInHandler())
	mux.HandleFunc("GET "+v.Config.Prefix+"oauth/logout", v.logOutHandler())
	mux.HandleFunc("POST "+v.Config.Prefix+"oauth/back_channel_logout", v.backChannelLogOutHandler())
}

func (v *OAuthValidator) Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		return func(h http.Handler) {
			v.challengeAction(w, r)
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
			v.challengeAction(w, r)
		}
	}

	_, err = v.Provider.Verifier(&oidc.Config{
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
			v.challengeAction(w, r)
		}
	}

	return func(h http.Handler) {
		r = r.WithContext(context.WithValue(r.Context(), ContextUserKey, sessionData))

		h.ServeHTTP(w, r)
	}
}

func (v *OAuthValidator) callbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := v.verifyState(r); err != nil {
			log.Printf("Error verifying state: %v", err)
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

		tokenClaims, err := v.validateAndGetClaimsIDToken(r.Context(), oauth2Token)
		if err != nil {
			log.Printf("Error validating and getting claims ID token: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessionID := (*tokenClaims)["sid"].(string)
		createdAt := util.TimeFromFloat64Seconds((*tokenClaims)["iat"].(float64))

		sessionData := session.SessionData{
			AccessToken:  oauth2Token.AccessToken,
			CreatedAt:    createdAt,
			Data:         *tokenClaims,
			RefreshToken: oauth2Token.RefreshToken,
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

func (v *OAuthValidator) challengeAction(w http.ResponseWriter, r *http.Request) {
	if v.Config.SignInOnChallenge {
		v.logInHandler()(w, r)
	} else {
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (v *OAuthValidator) logInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := fmt.Sprintf("%s://%s%s", util.GetScheme(r), r.Host, v.Config.Prefix+"oauth/callback")
		log.Printf("Redirect URL: %s", redirectURL)
		state := util.RandStringBytes(32)
		if err := v.StateStore.Set(r.Context(), state); err != nil {
			log.Printf("Error setting state: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		authURL := v.Oauth2Config.AuthCodeURL(
			state,
			oauth2.SetAuthURLParam("redirect_uri", redirectURL),
		)
		log.Printf("Auth URL: %s", authURL)
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

func (v *OAuthValidator) logOutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := r.Header.Get("Referer")
		if redirectURL == "" {
			redirectURL = fmt.Sprintf("%s://%s", util.GetScheme(r), r.Host)
		}
		log.Printf("Logout redirect URL: %s", redirectURL)
		logoutURL := fmt.Sprintf(
			"%s/realms/%s/protocol/openid-connect/logout?post_logout_redirect_uri=%s&client_id=%s",
			v.Config.AuthServerURL,
			url.PathEscape(v.Config.Realm),
			url.QueryEscape(redirectURL),
			v.Config.ClientID,
		)
		log.Printf("Logout URL: %s", logoutURL)
		http.SetCookie(w, &http.Cookie{
			Name:  "session_id",
			Value: "",
			Path:  "/",
		})
		http.Redirect(w, r, logoutURL, http.StatusTemporaryRedirect)
	}
}

func (v *OAuthValidator) backChannelLogOutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tokenString := r.FormValue("logout_token")
		if tokenString == "" {
			log.Printf("logout_token is required")
			http.Error(w, "logout_token is required", http.StatusBadRequest)
			return
		}

		logoutToken, err := v.Verifier.Verify(r.Context(), tokenString)
		if err != nil {
			log.Printf("Error verifying logout token: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var claims map[string]interface{}
		if err := logoutToken.Claims(&claims); err != nil {
			log.Printf("Error parsing logout token claims: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if v.Config.DumpClaims {
			log.Printf("Dump logout token claims: %+v", claims)
		}

		v.SessionStore.Delete(r.Context(), claims["sid"].(string))
	}
}

func (v *OAuthValidator) validateAndGetClaimsIDToken(ctx context.Context, oauth2Token *oauth2.Token) (*map[string]interface{}, error) {
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token is required")
	}

	idToken, err := v.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify id_token: %w", err)
	}

	claims := map[string]interface{}{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if v.Config.DumpClaims {
		log.Printf("Dump ID token claims: %+v", claims)
	}

	return &claims, nil
}

func (v *OAuthValidator) verifyState(r *http.Request) error {
	state := r.URL.Query().Get("state")

	if _, err := v.StateStore.Get(r.Context(), state); err != nil {
		return fmt.Errorf("invalid state: %v", err)
	}

	return nil
}
