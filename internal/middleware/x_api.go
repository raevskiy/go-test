package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(correctKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if correctKey == "" {
			ctx.Next()
		}

		apiKey := ctx.GetHeader("X-API-Key")

		if apiKey == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-API-Key header is missing"})
			return
		}

		if apiKey != correctKey {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}

		ctx.Next()
	}
}
