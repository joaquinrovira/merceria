package handler

import (
	"encoding/json"
	"merceria/internal/auth"
	"net/http"
)

// LogoutHandler clears the session cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearCookie(w, auth.SessionCookieName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Logged out successfully",
	})
}
