package auth

import "github.com/golang-jwt/jwt/v5"

type contextKey string

const UserIDKey contextKey = "user_id"

type Jwt struct {
	UserID int `json:"sub"`
	jwt.RegisteredClaims
}
