package handler

import (
	"log"
	"merceria/internal/auth"
	"net/http"
	"os"
)

func DevLoginHandler(rauth auth.RequestAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authr := rauth(r)
		claims := auth.NewSessionClaims("dev", "dev@example.com")
		claims = auth.SetSessionClaimsSpreadsheet(claims, os.Getenv("SPREADSHEET_ID"))
		claims.Dev = true

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

		w.WriteHeader(http.StatusOK)
	}
}
