package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"merceria/internal/auth"
	"net/http"
)

func LoginHandler(rauth auth.RequestAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth.InitiateLogin(rauth(r), w, r)
	}
}

func LoginCallbackHandler(rauth auth.RequestAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authr := rauth(r)

		// Get state from cookie
		stateCookie, err := r.Cookie(auth.OAuthStateCookieName)
		if err != nil {
			http.Error(w, "state cookie not found", http.StatusBadRequest)
			return
		}
		state := stateCookie.Value
		auth.ClearCookie(w, auth.OAuthStateCookieName)

		// Validate state
		if r.FormValue("state") != state {
			http.Error(w, "invalid state parameter", http.StatusUnauthorized)
			return
		}

		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "missing code parameter", http.StatusBadRequest)
			return
		}

		// Exchange authorization code for token
		token, err := authr.OauthConfig.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create HTTP client with token
		client := authr.OauthConfig.Client(r.Context(), token)

		// Fetch user info from Google
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, "failed to fetch user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Parse user info response
		var userInfo struct {
			ID            string `json:"id"`
			Email         string `json:"email"`
			VerifiedEmail bool   `json:"verified_email"`
			Name          string `json:"name"`
			GivenName     string `json:"given_name"`
			FamilyName    string `json:"family_name"`
			Picture       string `json:"picture"`
			Locale        string `json:"locale"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, "failed to parse user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if email is verified
		if !userInfo.VerifiedEmail {
			log.Printf("UNAUTHORIZED ACCESS: unverified email %s", userInfo.Email)
			http.Error(w, "email not verified", http.StatusForbidden)
			return
		}

		// Check if email is in the allow list
		if !authr.IsEmailAllowed(userInfo.Email) {
			log.Printf("UNAUTHORIZED ACCESS: email not in allow list %s", userInfo.Email)
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}

		// Log user information (be careful not to log sensitive data in production)
		log.Printf("User logged in: %s (%s)", userInfo.Email, userInfo.ID)

		// Create a JWT session token with embedded Google access token
		claims := auth.NewSessionClaims(userInfo.ID, userInfo.Email)

		// Optional:
		// Extract spreadsheet ID from cookie if it exists and include in new session token
		err = func() error {
			token, err := auth.GetSessionCookie(r)
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				return fmt.Errorf("failed to retrieve session cookie: %w", err)
			}
			if token == "" {
				return nil
			}
			prev, err := authr.DecodeSessionToken(token, false)
			if err != nil {
				return nil
			}
			if prev.SpreadsheetId == "" {
				return nil
			}
			claims = auth.SetSessionClaimsSpreadsheet(claims, prev.SpreadsheetId)
			return nil
		}()
		if err != nil {
			log.Printf("ERROR: failed to extract spreadsheet ID from previous session: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		sessionToken, err := authr.CreateSessionToken(claims)
		if err != nil {
			log.Printf("ERROR: failed to create session token: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		auth.SetSessionCookie(w, sessionToken)

		// If redirect URL is stored in cookie, redirect there
		redirectCookie, err := r.Cookie(auth.OAuthRedirectCookieName)
		if err == nil {
			auth.ClearCookie(w, auth.OAuthRedirectCookieName)
			redirectURL := redirectCookie.Value
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		// Otherwise, return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"user":    userInfo,
			"message": "Logged in successfully",
		})
	}
}
