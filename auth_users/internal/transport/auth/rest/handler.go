package rest

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth/dto"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"log/slog"
	"net/http"
)

type AuthService interface {
	Register(ctx context.Context, user *dto.RegisterRequest) error
	Login(ctx context.Context, user *dto.LoginRequest) (*dto.LoginResponse, error)
	VerifyEmail(ctx context.Context, token string) (*dto.RegisterResponse, error)
	ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
}

type AuthHandler struct {
	log         *slog.Logger
	authService AuthService
}

func NewAuthHandler(log *slog.Logger, authService AuthService) *AuthHandler {
	return &AuthHandler{
		log:         log,
		authService: authService,
	}
}

func (h *AuthHandler) RegisterRoutes(engine *gin.Engine) {
	engine.POST("/register", h.Register)
	engine.POST("/login", h.Login) // было GET — учётные данные не должны попадать в URL
	engine.GET("/verify", h.VerifyEmail)
	engine.POST("/forgot-password", h.ForgotPassword)
	engine.GET("/reset-password", h.ResetPasswordPage)
	engine.POST("/reset-password", h.ResetPasswordForm)
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var registerDTO dto.RegisterRequest

	if err := ctx.ShouldBindJSON(&registerDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid request body",
		})

		h.log.Error("registration failed", "error", err)

		return
	}

	err := h.authService.Register(ctx.Request.Context(), &registerDTO)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid input",
			})

		case errors.Is(err, core.ErrInvalidGender):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid gender",
			})

		case errors.Is(err, core.ErrInvalidCredentials):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid credentials",
			})

		case errors.Is(err, core.ErrInvalidBirthdate):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid birthdate",
			})

		case errors.Is(err, core.ErrTooYoung):
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "error",
				"error":  "too young",
			})

		case errors.Is(err, core.ErrPasswordTooShort):
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "error",
				"error":  "password too short",
			})

		case errors.Is(err, core.ErrRepeatWrongPassword):
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "error",
				"error":  "repeat password",
			})

		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "internal server error",
			})
		}

		h.log.Error("registration failed", "error", err)

		return
	}

	h.log.Info("registration succeeded")

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"result": "unconfirmed user created",
	})
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var loginDTO dto.LoginRequest

	if err := ctx.ShouldBindJSON(&loginDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid request body",
		})

		h.log.Error("login failed", "error", err)

		return
	}

	respData, err := h.authService.Login(ctx.Request.Context(), &loginDTO)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid input",
			})

		case errors.Is(err, core.ErrInvalidCredentials):
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "invalid credentials",
			})

		case errors.Is(err, core.ErrTooManyRequests):
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"status": "error",
				"error":  "too many requests",
			})

		case errors.Is(err, core.ErrTooManyAttempts):
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"status": "error",
				"error":  "too many failed attempts, try later",
			})

		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "internal server error",
			})
		}

		h.log.Error("login failed", "error", err)

		return
	}

	h.log.Info("login succeeded")

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   respData,
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token required",
		})
		return
	}

	_, err := h.authService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidToken):
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid token"})

		case errors.Is(err, core.ErrTokenExpired):
			c.JSON(http.StatusGone, gin.H{"error": "token expired"})

		case errors.Is(err, core.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "email verified",
	})
}

func (h *AuthHandler) ForgotPassword(ctx *gin.Context) {
	var req dto.ForgotPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid request body"})
		return
	}

	_ = h.authService.ForgotPassword(ctx.Request.Context(), &req)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"result": "if this email is registered, you will receive a reset link",
	})
}

func (h *AuthHandler) ResetPasswordPage(ctx *gin.Context) {
	token := ctx.Query("token")

	if token == "" {
		ctx.String(http.StatusBadRequest, "Invalid token")
		return
	}

	safeToken := html.EscapeString(token)

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Изменить пароль</title>
	</head>
	<body>
		<h2>Изменить пароль</h2>

		<form method="POST" action="/reset-password">
			<input type="hidden" name="token" value="%s"/>

			<input type="password" name="new_password" placeholder="Новый пароль" required minlength="6"/>
			<br><br>

			<input type="password" name="confirm_password" placeholder="Повторите пароль" required minlength="6"/>
			<br><br>

			<button type="submit">Подтвердить</button>
		</form>
	</body>
	</html>
	`, safeToken))
}

func (h *AuthHandler) ResetPasswordForm(ctx *gin.Context) {
	token := ctx.PostForm("token")
	newPassword := ctx.PostForm("new_password")
	confirmPassword := ctx.PostForm("confirm_password")

	if token == "" || newPassword == "" || confirmPassword == "" {
		ctx.String(http.StatusBadRequest, "Invalid request")
		return
	}

	if len(newPassword) < 6 {
		renderResetPage(ctx, token, "Пароль должен быть длиннее 6 символов", http.StatusUnprocessableEntity)
		return
	}

	if newPassword != confirmPassword {
		renderResetPage(ctx, token, "Пароли не совпадают", http.StatusUnprocessableEntity)
		return
	}

	err := h.authService.ResetPassword(ctx.Request.Context(), &dto.ResetPasswordRequest{
		Token:       token,
		NewPassword: newPassword,
	})

	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidToken):
			ctx.String(http.StatusBadRequest, "Invalid or expired token")
		case errors.Is(err, core.ErrTokenExpired):
			ctx.String(http.StatusBadRequest, "Token expired")
		default:
			ctx.String(http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	ctx.String(http.StatusOK, "Пароль успешно изменен!")
}

func renderResetPage(ctx *gin.Context, token string, errMsg string, status int) {
	safeToken := html.EscapeString(token)
	safeErr := html.EscapeString(errMsg)

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(status, fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Изменить пароль</title>
	</head>
	<body>
		<h2>Изменить пароль</h2>

		%s

		<form method="POST" action="/reset-password">
			<input type="hidden" name="token" value="%s"/>

			<input type="password" name="new_password" placeholder="Новый пароль" required minlength="6"/>
			<br><br>

			<input type="password" name="confirm_password" placeholder="Повторите пароль" required minlength="6"/>
			<br><br>

			<button type="submit">Подтвердить</button>
		</form>
	</body>
	</html>
	`,
		func() string {
			if safeErr != "" {
				return `<p style="color:red;">` + safeErr + `</p>`
			}
			return ""
		}(),
		safeToken,
	))
}
