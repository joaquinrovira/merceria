package auth

import (
	"net/http"

	"golang.org/x/oauth2"
)

func InitiateLogin(authr *Authorizer, w http.ResponseWriter, r *http.Request) {
	state := RandomState()
	http.SetCookie(w, NewStateCookie(state))
	url := authr.OauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
