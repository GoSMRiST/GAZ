package jwt

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"errors"
	"time"
)

type JwtInterface interface {
	CreateNewToken(userId int, duration time.Duration, secret string) (string, error)
	ParseJwtToken(tokenStr string, secret string) (*auth.Jwt, error)
}

type JwtService struct {
	secret   string
	duration time.Duration
	jwt      JwtInterface
}

func NewJwtService(secret string, duration time.Duration, jwt JwtInterface) *JwtService {
	return &JwtService{
		secret:   secret,
		duration: duration,
		jwt:      jwt,
	}
}

func (js *JwtService) CreateNewToken(userId int) (string, error) {
	if userId < 1 || js.duration == time.Duration(0) {
		return "", core.ErrInvalidInput
	}

	jwtToken, err := js.jwt.CreateNewToken(userId, js.duration, js.secret)
	if err != nil {
		return "", err
	}

	return jwtToken, nil
}

func (js *JwtService) ParseJwtToken(tokenStr string) (*auth.Jwt, error) {
	if tokenStr == "" {
		return nil, core.ErrInvalidToken
	}

	jwtToken, err := js.jwt.ParseJwtToken(tokenStr, js.secret)
	if err != nil {
		if errors.Is(err, core.ErrInvalidToken) {
			return nil, core.ErrInvalidToken
		}

		if errors.Is(err, core.ErrTokenExpired) {
			return nil, core.ErrTokenExpired
		}

		return nil, err
	}

	return jwtToken, nil
}
