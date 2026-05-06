package core

import (
	"auth_users/internal/core/auth"
	"context"
)

func GetUserID(ctx context.Context) (int, error) {
	val := ctx.Value(auth.UserIDKey)
	if val == nil {
		return 0, ErrUnauthorized
	}

	id, ok := val.(int)
	if !ok {
		return 0, ErrUnauthorized
	}

	return id, nil
}
