package rest

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"auth_users/internal/core/auth/dto"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) UpdateBio(ctx *gin.Context) {
	var req dto.UpdateBioRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid request body"})
		return
	}

	if err := h.userService.UpdateBio(ctx.Request.Context(), req.Bio); err != nil {
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "unauthorized"})
		case errors.Is(err, core.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "bio must be 300 characters or less"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "result": true})
}

func (h *UserHandler) UploadAvatar(ctx *gin.Context) {
	userID, ok := ctx.Request.Context().Value(auth.UserIDKey).(int)
	if !ok || userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "unauthorized"})
		return
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "no file provided"})
		return
	}

	// Ограничение 5 МБ
	if file.Size > 5*1024*1024 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "file too large, max 5MB"})
		return
	}

	dir := filepath.Join("..", "static", "avatars")
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.log.Error("UploadAvatar: mkdir failed", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "storage error"})
		return
	}

	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		ext = ".jpg"
	}

	filename := fmt.Sprintf("%d%s", userID, ext)
	savePath := filepath.Join(dir, filename)

	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		h.log.Error("UploadAvatar: save failed", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "failed to save file"})
		return
	}

	avatarPath := fmt.Sprintf("/static/avatars/%s?v=%d", filename, time.Now().Unix())

	if err := h.userService.UpdateAvatarURL(ctx.Request.Context(), &dto.UpdateAvatarRequest{
		NewAvatar: avatarPath,
	}); err != nil {
		h.log.Error("UploadAvatar: db update failed", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "failed to update avatar"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "avatar_url": avatarPath})
}
