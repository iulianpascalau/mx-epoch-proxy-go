package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenExpirationTime = time.Hour

// Claims struct holds the JWT claims
type Claims struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// Authenticator defines the behavior for authentication
type Authenticator interface {
	GenerateToken(username string, isAdmin bool) (string, error)
	CheckAuth(r *http.Request) (*Claims, error)
}

// JWTAuthenticator implements Authenticator using JWT
type JWTAuthenticator struct {
	jwtKey []byte
}

// NewJWTAuthenticator creates a new JWTAuthenticator
func NewJWTAuthenticator(key string) *JWTAuthenticator {
	return &JWTAuthenticator{
		jwtKey: []byte(key),
	}
}

// GenerateToken generates a new JWT token
func (ja *JWTAuthenticator) GenerateToken(username string, isAdmin bool) (string, error) {
	expirationTime := time.Now().Add(tokenExpirationTime)
	claims := &Claims{
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ja.jwtKey)
}

// ValidateToken validates the token string and returns claims
func (ja *JWTAuthenticator) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return ja.jwtKey, nil
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
func (ja *JWTAuthenticator) CheckAuth(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	return ja.ValidateToken(bearerToken[1])
}
