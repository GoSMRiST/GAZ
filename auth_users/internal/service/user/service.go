package user

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"auth_users/internal/core/auth/dto"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

var allowedStatuses = map[string]struct{}{
	"default": {},
	"vip":     {},
	"super":   {},
}

func IsValidStatus(s string) bool {
	_, ok := allowedStatuses[s]
	return ok
}

type RepositoryInterface interface {
	GetUserByID(ctx context.Context, id int) (*dto.GetByIdResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.AuthUser, error)
	UpdateNickname(ctx context.Context, id int, nickname *dto.UpdateNicknameRequest) error
	UpdateAvatarURL(ctx context.Context, id int, url *dto.UpdateAvatarRequest) error
	UpdatePassword(ctx context.Context, id int, newPasswordHash *dto.UpdatePasswordRequest) error
	UpdateSubscriptionStatus(ctx context.Context, id int, sub string) error
	UpdateBio(ctx context.Context, id int, bio string) error
}

type UserService struct {
	log  *slog.Logger
	repo RepositoryInterface
}

func NewUserService(log *slog.Logger, repo RepositoryInterface) *UserService {
	return &UserService{
		log:  log,
		repo: repo,
	}
}

func (s *UserService) GetUserByIDPublic(ctx context.Context, id int) (*dto.GetByIdResponse, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, core.ErrUserNotFound) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context) (*dto.GetByIdResponse, error) {
	id, ok := ctx.Value(auth.UserIDKey).(int)
	if !ok || id == 0 {
		return nil, core.ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get user", "error", err)

		if errors.Is(err, core.ErrUserNotFound) {
			return nil, core.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateNickname(ctx context.Context, nickname *dto.UpdateNicknameRequest) error {
	userId, err := core.GetUserID(ctx)
	if err != nil {
		s.log.Error("failed to get user id", "error", err)

		if errors.Is(err, core.ErrUnauthorized) {
			return core.ErrUnauthorized
		}

		return err
	}

	if nickname == nil || nickname.NewNickname == "" {
		s.log.Error("nickname is empty")

		return core.ErrNicknameIsEmpty
	}

	if err := s.repo.UpdateNickname(ctx, userId, nickname); err != nil {
		s.log.Error("failed to update user nickname:", "error", err)

		if errors.Is(err, core.ErrUserNotFound) {
			return core.ErrUserNotFound
		}

		return err
	}

	return nil
}

func (s *UserService) UpdateAvatarURL(ctx context.Context, url *dto.UpdateAvatarRequest) error {
	userId, err := core.GetUserID(ctx)
	if err != nil {
		s.log.Error("failed to get user id", "error", err)

		if errors.Is(err, core.ErrUnauthorized) {
			return core.ErrUnauthorized
		}

		return err
	}

	if url == nil || url.NewAvatar == "" {
		s.log.Error("url is empty")

		return core.ErrUrlIsEmpty
	}

	if err := s.repo.UpdateAvatarURL(ctx, userId, url); err != nil {
		s.log.Error("fail to update avatar url:", "error", err)

		if errors.Is(err, core.ErrUserNotFound) {
			return core.ErrUserNotFound
		}

		return err
	}

	return nil
}

func (s *UserService) UpdatePassword(ctx context.Context, newPassword *dto.UpdatePasswordRequest) error {
	userId, err := core.GetUserID(ctx)
	if err != nil {
		s.log.Error("failed to get user id", "error", err)

		if errors.Is(err, core.ErrUnauthorized) {
			return core.ErrUnauthorized
		}

		return err
	}

	if newPassword == nil || newPassword.NewPassword == "" {
		s.log.Error("newPassword is empty")

		return core.ErrPasswordIsEmpty
	}

	userByID, err := s.repo.GetUserByID(ctx, userId)
	if err != nil {
		return err
	}

	userByEmail, err := s.repo.GetUserByEmail(ctx, userByID.Email)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(userByEmail.PasswordHash),
		[]byte(newPassword.OldPassword),
	)
	if err != nil {
		return core.ErrWrongPassword
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newPassword.NewPassword = string(newPasswordHash)

	if err := s.repo.UpdatePassword(ctx, userId, newPassword); err != nil {
		s.log.Error("fail to update password:", "error", err)

		if errors.Is(err, core.ErrUserNotFound) {
			return core.ErrUserNotFound
		}

		return err
	}

	return nil
}

func (s *UserService) UpdateBio(ctx context.Context, bio string) error {
	if len([]rune(bio)) > 300 {
		return core.ErrInvalidInput
	}
	userId, err := core.GetUserID(ctx)
	if err != nil {
		return core.ErrUnauthorized
	}
	return s.repo.UpdateBio(ctx, userId, bio)
}

func (s *UserService) UpdateSubscription(ctx context.Context, sub *dto.UpdateSubscriptionRequest) error {
	if sub == nil || sub.SubscriptionStatus == "" {
		s.log.Error("subscription is empty")
		return core.ErrInvalidInput
	}

	if ok := IsValidStatus(sub.SubscriptionStatus); ok == false {
		s.log.Error("invalid subscription status", "subscription_status", sub.SubscriptionStatus)
		return core.ErrWrongSubStatus
	}

	userId, err := core.GetUserID(ctx)
	if err != nil {
		s.log.Error("failed to get user id", "error", err)
		if errors.Is(err, core.ErrUnauthorized) {
			return core.ErrUnauthorized
		}

		return err
	}

	if err := s.repo.UpdateSubscriptionStatus(ctx, userId, sub.SubscriptionStatus); err != nil {
		s.log.Error("fail to update subscription status:", "error", err)

		if errors.Is(err, core.ErrUserNotFound) {
			return core.ErrUserNotFound
		}

		return err
	}

	return nil
}
