package dto

type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateNicknameRequest struct {
	NewNickname string `json:"new_nickname" binding:"required"`
}

type UpdateAvatarRequest struct {
	NewAvatar string `json:"new_avatar" binding:"required"`
}

type UpdateSubscriptionRequest struct {
	SubscriptionStatus string `json:"subscription_id" binding:"required"`
}
