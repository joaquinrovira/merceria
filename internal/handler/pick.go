package handler

import (
	"context"
	"encoding/json"
	"merceria/internal/auth"
	"merceria/internal/middleware"
	"merceria/internal/util"
	"net/http"
	"os"
	"text/template"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func PickHandler(ctx context.Context, rauth auth.RequestAuthorizer, fs *os.Root, apiKey string) http.HandlerFunc {
	const name = "picker.html"
	data := ReloadingFile(ctx, fs, name)
	return func(w http.ResponseWriter, r *http.Request) {
		authr := rauth(r)
		sessionToken, err := auth.GetSessionCookie(r)
		if err != nil || sessionToken == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := authr.ValidateSessionToken(sessionToken)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		t := util.Must(template.New("").Parse(string(util.Must(data()))))
		m := map[string]interface{}{
			"Email":    claims.Email,
			"ClientID": authr.OauthConfig.ClientID,
			"APIKey":   apiKey,
			"APPId":    "801259244205",
			"Token":    "",
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t.Execute(w, m)
	}
}

func PickCallback(ctx context.Context, rauth auth.RequestAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authr := rauth(r)

		ctx := r.Context()
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		id := r.FormValue("file-id")
		name := r.FormValue("file-name")
		token := r.FormValue("token")

		if id == "" {
			http.Error(w, "spreadsheet_id is required", http.StatusBadRequest)
			return
		}

		d, err := drive.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))
		if err != nil {
			http.Error(w, "failed to create drive service: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = d.Permissions.Create(id, &drive.Permission{
			Type:         "user",
			Role:         "writer",
			EmailAddress: os.Getenv("GOOGLE_CLIENT_EMAIL"),
		}).Context(ctx).SendNotificationEmail(false).Do()
		if err != nil {
			http.Error(w, "failed to add permission: "+err.Error(), http.StatusInternalServerError)
			return
		}

		claims, ok := r.Context().Value(middleware.JwtClaims).(*auth.SessionClaims)
		if !ok || claims == nil {
			http.Error(w, "unauthorized: no session claims found", http.StatusUnauthorized)
			return
		}

		newClaims := auth.SetSessionClaimsSpreadsheet(claims, id)
		sessionToken, err := authr.CreateSessionToken(newClaims)
		if err != nil {
			http.Error(w, "failed to create session token with spreadsheet id: "+err.Error(), http.StatusInternalServerError)
			return
		}
		auth.SetSessionCookie(w, sessionToken)

		// If redirect URL is stored in cookie, redirect there
		redirectCookie, err := r.Cookie(auth.OAuthRedirectCookieName)
		if err == nil {
			auth.ClearCookie(w, auth.OAuthRedirectCookieName)
			redirectURL := redirectCookie.Value
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":           "ok",
			"spreadsheet_id":   id,
			"spreadsheet_name": name,
		})
	}
}
