package dto

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"       binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
