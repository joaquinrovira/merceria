package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/sheets/v4"
)

type Config struct {
	// Debug
	Development bool

	// Infra
	Port        string
	CORSOrigins []string

	// Auth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleAPIKey       string
	SessionSecret      string
	AllowedUsers       []string

	// Spreadsheet Service
	ServiceAccount *jwt.Config
	SpreadsheetID  string
}

func New() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	originsStr := os.Getenv("CORS_ORIGINS")
	var origins []string
	if originsStr != "" {
		origins = strings.Split(originsStr, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
	}

	allowedUsersStr := os.Getenv("ALLOWED_USERS")
	var allowedUsers []string
	if allowedUsersStr != "" {
		allowedUsers = strings.Split(allowedUsersStr, ",")
		for i := range allowedUsers {
			allowedUsers[i] = strings.TrimSpace(strings.ToLower(allowedUsers[i]))
		}
	}

	data, err := base64.StdEncoding.DecodeString(os.Getenv("SERVICE_ACCOUNT_JSON_BASE64"))
	if err != nil {
		return nil, fmt.Errorf("decode SERVICE_ACCOUNT_JSON_BASE64 credentials: %w", err)
	}

	jwt, err := google.JWTConfigFromJSON(data, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	return &Config{
		Development: os.Getenv("DEVELOPMENT") != "" || os.Getenv("DEV") != "",

		Port:        port,
		CORSOrigins: origins,

		AllowedUsers:       allowedUsers,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleAPIKey:       os.Getenv("GOOGLE_API_KEY"),
		SessionSecret:      os.Getenv("SESSION_SECRET"),

		ServiceAccount: jwt,
		SpreadsheetID:  os.Getenv("SPREADSHEET_ID"),
	}, nil
}
