package dto

type AuthUser struct {
	ID           int    `json:"user_id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	IsVerified   bool   `json:"-"`
}
