package auth

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTTL         = 28 * time.Hour
	RefreshThreshold = 12 * time.Hour
)

type SessionClaims struct {
	Email         string `json:"email"`
	SpreadsheetId string `json:"sid,omitzero"`
	Dev           bool   `json:"dev,omitzero"`
	jwt.RegisteredClaims
}

func NewSessionClaims(sub, email string) *SessionClaims {
	now := time.Now()
	claims := &SessionClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		},
	}
	return claims
}

func SetSessionClaimsSpreadsheet(claims *SessionClaims, id string) *SessionClaims {
	next := *claims
	next.SpreadsheetId = id
	return &next
}

func RefreshSessionClaims(claims *SessionClaims) *SessionClaims {
	now := time.Now()
	next := &SessionClaims{
		Email: claims.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
			Audience:  slices.Clone(claims.Audience),
		},
	}
	return next
}

func (a *Authorizer) CreateSessionToken(claims *SessionClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	v, err := token.SignedString([]byte(a.SessionSecret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return v, err
}

func (a *Authorizer) ValidateSessionToken(tokenString string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.SessionSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
