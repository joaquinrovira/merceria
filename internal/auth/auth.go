package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"merceria/internal/config"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const SessionCookieName = "merceria_session"

// RandomState generates a random state string for CSRF protection
func RandomState() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func OauthConfig(location string) *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  location + "/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/drive.file",
		},
		Endpoint: google.Endpoint,
	}
}

type RequestAuthorizer func(*http.Request) *Authorizer

func NewAuthorizerFactory(cfg *config.Config) RequestAuthorizer {
	return func(r *http.Request) *Authorizer {
		// Get source location for constructing RedirectURL

		host := r.Host
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if !cfg.Development { // TODO: fix this for production behind a reverse proxy
			scheme = "https"
		}

		location := fmt.Sprintf("%s://%s", scheme, host)

		return &Authorizer{
			AllowedUsers:  cfg.AllowedUsers,
			SessionSecret: cfg.SessionSecret,
			OauthConfig:   OauthConfig(location),
		}
	}
}

type Authorizer struct {
	AllowedUsers  []string
	SessionSecret string
	OauthConfig   *oauth2.Config
}

func (a *Authorizer) IsEmailAllowed(email string) bool {
	if len(a.AllowedUsers) == 0 {
		return false
	}
	emailLower := strings.ToLower(email)
	for _, allowed := range a.AllowedUsers {
		if allowed == emailLower {
			return true
		}
	}
	return false
}
