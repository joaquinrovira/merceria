package middleware

import (
	"log"
	"net/http"

	"merceria/internal/auth"
)

func WithSpreadsheetId(rauth auth.RequestAuthorizer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(JwtClaims).(*auth.SessionClaims)
			if !ok || claims == nil {
				log.Printf("UNAUTHORIZED: no JWT claims found in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.SpreadsheetId == "" {
				location := r.URL.RequestURI()
				http.SetCookie(w, auth.NewRedirectCookie(location))
				http.Redirect(w, r, "/pick", http.StatusTemporaryRedirect)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
