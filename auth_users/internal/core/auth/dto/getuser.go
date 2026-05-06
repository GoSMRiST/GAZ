package dto

import "time"

type GetByIdRequest struct {
	UserId int `json:"user_id"`
}

type GetByIdResponse struct {
	ID            int       `json:"user_id"`
	Nickname      string    `json:"nickname"`
	Email         string    `json:"email"`
	BirthDate     time.Time `json:"birth_date"`
	Gender        string    `json:"gender"`
	AvatarURL     string    `json:"avatar_url"`
	MeetingsCount int       `json:"meetings_count"`
	IsVerified    bool      `json:"is_verified"`
	Subscription  string    `json:"subscription"`
	Bio           string    `json:"bio"`
}

// UpdateBioRequest — тело запроса PATCH /bio.
type UpdateBioRequest struct {
	Bio string `json:"bio"`
}
