package auth

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"auth_users/internal/core/auth/dto"
	"auth_users/internal/infrastructure/redis"
	"auth_users/internal/util"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

const loginDelay = 500 * time.Millisecond

const (
	man   = "man"
	woman = "woman"
)

type RepositoryInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*dto.AuthUser, error)
	GetUserByID(ctx context.Context, id int) (*dto.GetByIdResponse, error)
	CreateEmailVerification(ctx context.Context, req *dto.RegisterRequest, token string, expiresAt time.Time) error
	GetVerificationByToken(ctx context.Context, token string) (*auth.EmailVerification, error)
	DeleteVerificationByToken(ctx context.Context, token string) error
	CreateUser(ctx context.Context, v *auth.EmailVerification) (*dto.RegisterResponse, error)
	CreatePasswordReset(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetPasswordResetByToken(ctx context.Context, token string) (int, time.Time, error)
	DeletePasswordReset(ctx context.Context, token string) error
	UpdatePasswordByID(ctx context.Context, userID int, passwordHash string) error
}

type JwtInterface interface {
	CreateNewToken(userId int) (string, error)
	ParseJwtToken(tokenStr string) (*auth.Jwt, error)
}

type EmailSender interface {
	SendVerification(to string, token string) error
	SendPasswordReset(to string, token string) error
}

type AuthService struct {
	log        *slog.Logger
	repo       RepositoryInterface
	jwtService JwtInterface
	redis      *redis.Client
	email      EmailSender
}

func NewAuthService(log *slog.Logger, repo RepositoryInterface, jwtService JwtInterface, redis *redis.Client, email EmailSender) *AuthService {
	return &AuthService{
		log:        log,
		repo:       repo,
		jwtService: jwtService,
		redis:      redis,
		email:      email,
	}
}

func (s *AuthService) rateLimitLogin(ctx context.Context, email string) error {
	key := "rl:login:" + email

	ok, err := s.redis.Rdb.SetNX(ctx, key, 1, 3*time.Second).Result()
	if err != nil {
		return err
	}

	if !ok {
		return core.ErrTooManyRequests
	}

	return nil
}

func (s *AuthService) incLoginFail(ctx context.Context, email string) error {
	key := "bf:login:" + email

	pipe := s.redis.Rdb.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 10*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	if incr.Val() >= 5 {
		return core.ErrTooManyAttempts
	}

	return nil
}

func (s *AuthService) resetLoginFail(ctx context.Context, email string) {
	key := "bf:login:" + email
	if err := s.redis.Rdb.Del(ctx, key).Err(); err != nil {
		// Некритично — счётчик просто протухнет через 10 минут сам.
		// Логируем чтобы заметить проблемы с Redis.
		s.log.Warn("resetLoginFail: failed to delete brute-force key", "key", key, "error", err)
	}
}

func (s *AuthService) Register(ctx context.Context, user *dto.RegisterRequest) error {
	if user.Email == "" || user.Password == "" || user.Nickname == "" || user.Birthdate == "" {
		return core.ErrInvalidInput
	}

	if len([]rune(user.Password)) <= 6 {
		return core.ErrPasswordTooShort
	}

	if user.Password != user.RepeatPassword {
		return core.ErrRepeatWrongPassword
	}

	if user.Gender != man && user.Gender != woman {
		return core.ErrInvalidGender
	}

	birthDate, err := time.Parse("2006-01-02", user.Birthdate)
	if err != nil {
		return core.ErrInvalidBirthdate
	}
	if time.Since(birthDate) < 14*365*24*time.Hour {
		return core.ErrTooYoung
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)

	_, err = s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		if !errors.Is(err, core.ErrUserNotFound) {
			return err
		}
	} else {
		return core.ErrInvalidCredentials
	}

	token, err := util.GenerateToken(32)
	if err != nil {
		return err
	}

	expires := time.Now().Add(30 * time.Minute)

	err = s.repo.CreateEmailVerification(ctx, user, token, expires)
	if err != nil {
		return err
	}

	if err := s.email.SendVerification(user.Email, token); err != nil {
		s.log.Error("Register: failed to send verification email", "err", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, user *dto.LoginRequest) (*dto.LoginResponse, error) {
	if user.Email == "" || user.Password == "" {
		s.log.Info("invalid input")
		return nil, core.ErrInvalidInput
	}

	if err := s.rateLimitLogin(ctx, user.Email); err != nil {
		return nil, err
	}

	dbUser, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		s.log.Info("user not found", "error", err)

		_ = s.incLoginFail(ctx, user.Email)
		time.Sleep(loginDelay)

		if errors.Is(err, core.ErrUserNotFound) {
			return nil, core.ErrInvalidCredentials
		}

		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(dbUser.PasswordHash),
		[]byte(user.Password),
	)
	if err != nil {
		s.log.Warn("login failed", "email", user.Email)

		_ = s.incLoginFail(ctx, user.Email)
		time.Sleep(500 * time.Millisecond)

		return nil, core.ErrInvalidCredentials
	}

	s.resetLoginFail(ctx, user.Email)

	token, err := s.jwtService.CreateNewToken(dbUser.ID)
	if err != nil {
		if errors.Is(err, core.ErrInvalidInput) {
			s.log.Info("invalid input", "error", err)
			return nil, core.ErrInvalidInput
		}

		return nil, err
	}

	return &dto.LoginResponse{Token: token}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) (*dto.RegisterResponse, error) {
	data, err := s.repo.GetVerificationByToken(ctx, token)
	if err != nil {
		return nil, core.ErrInvalidToken
	}

	if time.Now().After(data.Expires) {
		// Токен просрочен — удаляем его, пользователь должен зарегистрироваться заново
		_ = s.repo.DeleteVerificationByToken(ctx, token)
		return nil, core.ErrTokenExpired
	}

	user, err := s.repo.CreateUser(ctx, data)
	if err != nil {
		// Не удаляем токен — пользователь сможет повторить попытку
		s.log.Error("VerifyEmail: failed to create user", "error", err)
		return nil, err
	}

	// Удаляем токен только после успешного создания пользователя
	if err := s.repo.DeleteVerificationByToken(ctx, token); err != nil {
		// Некритично: cleanup worker уберёт его позже, пользователь уже создан
		s.log.Warn("VerifyEmail: failed to delete verification token", "error", err)
	}

	return user, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	if req.Email == "" {
		return core.ErrInvalidInput
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.log.Info("forgot password: user not found or error", "err", err)
		return nil
	}

	token, err := util.GenerateToken(32)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	if err := s.repo.CreatePasswordReset(ctx, user.ID, token, expiresAt); err != nil {
		return err
	}

	if err := s.email.SendPasswordReset(user.Email, token); err != nil {
		s.log.Error("forgot password: failed to send email", "err", err)
		// Не возвращаем ошибку — токен уже создан, пользователь может повторить
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	if req.Token == "" || req.NewPassword == "" {
		return core.ErrInvalidInput
	}

	userID, expiresAt, err := s.repo.GetPasswordResetByToken(ctx, req.Token)
	if err != nil {
		return core.ErrInvalidToken
	}

	if time.Now().After(expiresAt) {
		_ = s.repo.DeletePasswordReset(ctx, req.Token)
		return core.ErrTokenExpired
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePasswordByID(ctx, userID, string(hash)); err != nil {
		return err
	}

	_ = s.repo.DeletePasswordReset(ctx, req.Token)

	return nil
}
