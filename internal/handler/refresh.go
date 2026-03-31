package handler

import (
	"fmt"
	"merceria/internal/auth"
	"merceria/internal/util"
	tokencache "merceria/internal/util/token_cache.go"
	"net/http"
)

func RefreshHandler(rauth auth.RequestAuthorizer, tokenCache tokencache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authr := rauth(r)
		session := util.Must(auth.GetSessionCookie(r))
		claims := util.Must(authr.ValidateSessionToken(session))
		token := util.Must(tokenCache.Get(claims.Subject).Value().Token()).AccessToken
		fmt.Println(token)
		w.WriteHeader(http.StatusOK)
	}
}
