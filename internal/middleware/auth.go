package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"merceria/internal/auth"
)

type contextKey string

const JwtClaims contextKey = "jwt_claims"

func Auth(rauth auth.RequestAuthorizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authr := rauth(r)

			claims, err := checkCookie(authr, w, r)
			if err == nil {
				ctx := context.WithValue(r.Context(), JwtClaims, claims)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			if r.Method == http.MethodGet {
				location := r.URL.RequestURI()
				http.SetCookie(w, auth.NewRedirectCookie(location))
				auth.InitiateLogin(authr, w, r)
				return
			}

			log.Printf("UNAUTHORIZED: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}
}

func checkCookie(authr *auth.Authorizer, w http.ResponseWriter, r *http.Request) (*auth.SessionClaims, error) {
	sessionToken, err := auth.GetSessionCookie(r)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session cookie: %w", err)
	}
	if sessionToken == "" {
		return nil, fmt.Errorf("no session cookie found")
	}

	claims, err := authr.ValidateSessionToken(sessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Refresh session JWT if past threshold to extend session duration
	shouldRefresh := claims.IssuedAt != nil && time.Since(claims.IssuedAt.Time) > auth.RefreshThreshold
	if shouldRefresh {
		newToken, err := authr.CreateSessionToken(claims)
		if err != nil {
			log.Printf("ERROR: failed to refresh session token for user %s: %v", claims.Email, err)
		} else {
			auth.SetSessionCookie(w, newToken)
		}
	}

	return claims, nil
}
