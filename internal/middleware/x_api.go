package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func APIKeyAuth(correctKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if correctKey != "" {
			apiKey := ctx.GetHeader("X-API-Key")

			if apiKey == "" {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-API-Key header is missing"})
				return
			}

			if apiKey != correctKey {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
				return
			}
		}

		ctx.Next()
	}
}
