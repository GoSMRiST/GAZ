package dto

import "time"

type RegisterRequest struct {
	Nickname       string `json:"nickname" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	RepeatPassword string `json:"repeat_password" binding:"required,min=6"`
	Birthdate      string `json:"birthdate"`
	Gender         string `json:"gender"`
}

type RegisterResponse struct {
	ID         int       `json:"id"`
	Nickname   string    `json:"nickname"`
	Email      string    `json:"email"`
	Birthdate  time.Time `json:"birthdate"`
	Gender     string    `json:"gender,omitempty"`
	IsVerified bool      `json:"is_verified"`
}
