package jwt

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"
)

type JwtToken struct{}

func NewJwtToken() *JwtToken {
	return &JwtToken{}
}

func (jt *JwtToken) CreateNewToken(userId int, duration time.Duration, secret string) (string, error) {
	claims := auth.Jwt{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (jt *JwtToken) ParseJwtToken(tokenStr string, secret string) (*auth.Jwt, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &auth.Jwt{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, core.ErrTokenExpired
		}

		return nil, core.ErrInvalidToken
	}

	claims, ok := token.Claims.(*auth.Jwt)
	if !ok || !token.Valid {
		return nil, core.ErrInvalidToken
	}

	return claims, nil
}
