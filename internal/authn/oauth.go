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
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/store/state"
	"kdex.dev/proxy/internal/util"
)

type OAuthValidator struct {
	Config            *config.AuthnConfig
	Oauth2Config      *oauth2.Config
	Provider          *oidc.Provider
	SessionCookieName string
	SessionStore      *session.SessionStore
	StateStore        *state.StateStore
	Verifier          *oidc.IDTokenVerifier
}

func NewOAuthValidator(
	config *config.Config,
) *OAuthValidator {
	providerURL := fmt.Sprintf("%s/realms/%s", config.Authn.OAuth.AuthServerURL, url.PathEscape(config.Authn.Realm))
	provider, err := oidc.NewProvider(context.Background(), providerURL)

	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	scopes := config.Authn.OAuth.Scopes
	if !slices.Contains(scopes, oidc.ScopeOpenID) {
		scopes = append(scopes, oidc.ScopeOpenID)
	}
	if !slices.Contains(scopes, "email") {
		scopes = append(scopes, "email")
	}
	if !slices.Contains(scopes, "profile") {
		scopes = append(scopes, "profile")
	}
	if !slices.Contains(scopes, "roles") {
		scopes = append(scopes, "roles")
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.Authn.OAuth.ClientID,
	})

	oauth2Config := oauth2.Config{
		ClientID:     config.Authn.OAuth.ClientID,
		ClientSecret: config.Authn.OAuth.ClientSecret,
		RedirectURL:  config.Authn.OAuth.RedirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	stateStore, err := state.NewStateStore(config)
	if err != nil {
		log.Fatalf("Failed to create state store: %v", err)
	}

	sessionStore, err := session.NewSessionStore(config)
	if err != nil {
		log.Fatalf("Failed to create session store: %v", err)
	}

	return &OAuthValidator{
		Config:            &config.Authn,
		Oauth2Config:      &oauth2Config,
		Provider:          provider,
		SessionCookieName: config.Session.CookieName,
		SessionStore:      &sessionStore,
		StateStore:        &stateStore,
		Verifier:          verifier,
	}
}

func (v *OAuthValidator) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+v.Config.OAuth.Prefix+"/callback", v.callbackHandler())
	mux.HandleFunc("GET "+v.Config.OAuth.Prefix+"/login", v.logInHandler())
	mux.HandleFunc("GET "+v.Config.OAuth.Prefix+"/logout", v.logOutHandler())
	mux.HandleFunc("POST "+v.Config.OAuth.Prefix+"/back_channel_logout", v.backChannelLogOutHandler())
}

func (v *OAuthValidator) Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler) {
	sessionCookie, err := r.Cookie(v.SessionCookieName)
	if err != nil {
		return nil
	}

	sessionData, err := (*v.SessionStore).Get(r.Context(), sessionCookie.Value)
	if err != nil {
		return func(h http.Handler) {
			http.SetCookie(w, &http.Cookie{
				Name:  v.SessionCookieName,
				Value: "",
				Path:  "/",
			})
			h.ServeHTTP(w, r)
		}
	}

	_, err = v.Provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true, // ?
	}).Verify(r.Context(), sessionData.AccessToken)

	if err != nil {
		return func(h http.Handler) {
			(*v.SessionStore).Delete(r.Context(), sessionCookie.Value)
			http.SetCookie(w, &http.Cookie{
				Name:  v.SessionCookieName,
				Value: "",
				Path:  "/",
			})
			v.challengeAction(w, r)
		}
	}

	return func(h http.Handler) {
		r = r.WithContext(context.WithValue(r.Context(), kctx.SessionDataKey, sessionData))

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

		if err := (*v.SessionStore).Set(r.Context(), sessionID, sessionData); err != nil {
			log.Printf("Error setting session: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Name:     v.SessionCookieName,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Secure:   util.GetScheme(r) == "https",
			Value:    sessionID,
		})

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func (v *OAuthValidator) challengeAction(w http.ResponseWriter, r *http.Request) {
	if v.Config.OAuth.SignInOnChallenge {
		v.logInHandler()(w, r)
	} else {
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (v *OAuthValidator) logInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := fmt.Sprintf("%s://%s%s", util.GetScheme(r), r.Host, v.Config.OAuth.Prefix+"/callback")
		state := util.RandStringBytes(32)
		if err := (*v.StateStore).Set(r.Context(), state); err != nil {
			log.Printf("Error setting state: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		authURL := v.Oauth2Config.AuthCodeURL(
			state,
			oauth2.SetAuthURLParam("redirect_uri", redirectURL),
		)
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

func (v *OAuthValidator) logOutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := fmt.Sprintf("%s://%s", util.GetScheme(r), r.Host)
		logoutURL := fmt.Sprintf(
			"%s/realms/%s/protocol/openid-connect/logout?post_logout_redirect_uri=%s&client_id=%s",
			v.Config.OAuth.AuthServerURL,
			url.PathEscape(v.Config.Realm),
			url.QueryEscape(redirectURL),
			v.Config.OAuth.ClientID,
		)
		http.SetCookie(w, &http.Cookie{
			Name:  v.SessionCookieName,
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

		if v.Config.OAuth.DumpClaims {
			log.Printf("Dump logout token claims: %+v", claims)
		}

		(*v.SessionStore).Delete(r.Context(), claims["sid"].(string))
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

	if v.Config.OAuth.DumpClaims {
		log.Printf("Dump ID token claims: %+v", claims)
	}

	return &claims, nil
}

func (v *OAuthValidator) verifyState(r *http.Request) error {
	state := r.URL.Query().Get("state")

	if _, err := (*v.StateStore).Get(r.Context(), state); err != nil {
		return fmt.Errorf("invalid state: %v", err)
	}

	return nil
}
