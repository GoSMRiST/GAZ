package middleware

import (
	"auth_users/internal/core/auth"
	"context"
	"github.com/gin-gonic/gin"
)

type JwtInterface interface {
	ParseJwtToken(tokenStr string) (*auth.Jwt, error)
}

func AuthMiddleware(jwtService JwtInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}

		const prefix = "Bearer "
		if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token format"})
			return
		}

		tokenStr := authHeader[len(prefix):]

		claims, err := jwtService.ParseJwtToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		ctx := context.WithValue(c.Request.Context(), auth.UserIDKey, claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
