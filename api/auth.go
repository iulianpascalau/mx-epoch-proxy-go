package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

// SetJwtKey sets the key used for signing tokens
func SetJwtKey(key string) {
	jwtKey = []byte(key)
}

// Claims defines the JWT claims
type Claims struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token
func GenerateToken(username string, isAdmin bool) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateToken validates the token string and returns claims
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// CheckAuth validates the token from Authorization header.
// It returns claims if valid, error otherwise.
func CheckAuth(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	return ValidateToken(bearerToken[1])
}
