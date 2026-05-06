package core

import "errors"

var (
	// General errors
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input data")

	// Token errors
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")
	ErrUnauthorized = errors.New("unauthorized")

	// Users errors
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidGender      = errors.New("invalid gender")

	// Redis
	ErrTooManyRequests = errors.New("too many requests")
	ErrTooManyAttempts = errors.New("too many attempts")

	// Update errors
	ErrNicknameIsEmpty     = errors.New("nickname is empty")
	ErrUrlIsEmpty          = errors.New("url is empty")
	ErrPasswordIsEmpty     = errors.New("password is empty")
	ErrWrongPassword       = errors.New("wrong password")
	ErrRepeatWrongPassword = errors.New("repeat password")
	ErrWrongSubStatus      = errors.New("wrong subscription status")

	// Registration validation
	ErrPasswordTooShort = errors.New("password must be longer than 6 characters")
	ErrTooYoung         = errors.New("you must be at least 14 years old")
	ErrInvalidBirthdate = errors.New("invalid birthdate format, use YYYY-MM-DD")
)
