package rest

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth/dto"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type UserService interface {
	GetUserByID(ctx context.Context) (*dto.GetByIdResponse, error)
	GetUserByIDPublic(ctx context.Context, id int) (*dto.GetByIdResponse, error)
	UpdateNickname(ctx context.Context, nickname *dto.UpdateNicknameRequest) error
	UpdateAvatarURL(ctx context.Context, url *dto.UpdateAvatarRequest) error
	UpdatePassword(ctx context.Context, newPassword *dto.UpdatePasswordRequest) error
	UpdateBio(ctx context.Context, bio string) error
}

type UserHandler struct {
	log         *slog.Logger
	userService UserService
}

func NewUserHandler(log *slog.Logger, userService UserService) *UserHandler {
	return &UserHandler{
		log:         log,
		userService: userService,
	}
}

func (h *UserHandler) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/user/me", h.GetUser)
	engine.GET("/user/:id", h.GetUserByID)
	engine.PATCH("/nickname", h.UpdateNickname)
	engine.PATCH("/avatar", h.UpdateAvatarURL)
	engine.PATCH("/password", h.UpdatePassword)
	engine.PATCH("/bio", h.UpdateBio)
	engine.POST("/avatar-upload", h.UploadAvatar)
}

func (h *UserHandler) GetUser(ctx *gin.Context) {
	resp, err := h.userService.GetUserByID(ctx.Request.Context())
	if err != nil {
		switch {
		case errors.Is(err, core.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid request body",
			})

		case errors.Is(err, core.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "user not found",
			})

		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "internal server error",
			})
		}

		h.log.Error("getUser failed", "error", err)

		return
	}

	h.log.Debug("getUser succeed", "id", resp.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
	})
}

func (h *UserHandler) GetUserByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid user id"})
		return
	}

	resp, err := h.userService.GetUserByIDPublic(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, core.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "data": resp})
}

func (h *UserHandler) UpdateNickname(ctx *gin.Context) {
	var updateNickname dto.UpdateNicknameRequest

	if err := ctx.ShouldBindJSON(&updateNickname); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid request body",
		})

		h.log.Error("updateNickname failed", "error", err)

		return
	}

	err := h.userService.UpdateNickname(ctx.Request.Context(), &updateNickname)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "unauthorized",
			})

		case errors.Is(err, core.ErrNicknameIsEmpty):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid request body",
			})

		case errors.Is(err, core.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "user not found",
			})

		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "internal server error",
			})
		}

		h.log.Error("updateNickname failed", "error", err)

		return
	}

	h.log.Info("update succeeded", "result", true)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"result": true,
	})
}

func (h *UserHandler) UpdateAvatarURL(ctx *gin.Context) {
	var updateAvatarURL dto.UpdateAvatarRequest

	if err := ctx.ShouldBindJSON(&updateAvatarURL); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid request body",
		})

		h.log.Error("updateAvatarURL failed", "error", err)
		return
	}

	err := h.userService.UpdateAvatarURL(ctx.Request.Context(), &updateAvatarURL)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "unauthorized",
			})

		case errors.Is(err, core.ErrUrlIsEmpty):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid request body",
			})

		case errors.Is(err, core.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "user not found",
			})

		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "internal server error",
			})
		}

		h.log.Error("updateAvatarURL failed", "error", err)

		return
	}

	h.log.Info("update succeeded", "result", true)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"result": true,
	})
}

func (h *UserHandler) UpdatePassword(ctx *gin.Context) {
	var updatePassword dto.UpdatePasswordRequest

	if err := ctx.ShouldBindJSON(&updatePassword); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "invalid request body",
		})

		h.log.Error("updatePassword failed", "error", err)

		return
	}

	err := h.userService.UpdatePassword(ctx.Request.Context(), &updatePassword)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "unauthorized",
			})

		case errors.Is(err, core.ErrPasswordIsEmpty):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid request body",
			})

		case errors.Is(err, core.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "user not found",
			})

		case errors.Is(err, core.ErrWrongPassword):
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "wrong password",
			})

		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "internal server error",
			})
		}

		h.log.Error("updatePassword failed", "error", err)

		return
	}

	h.log.Info("update succeeded", "result", true)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"result": true,
	})
}
