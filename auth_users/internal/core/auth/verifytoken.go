package auth

import "time"

type EmailVerification struct {
	Email     string
	Nickname  string
	Password  string
	BirthDate time.Time
	Gender    string
	Expires   time.Time
}
