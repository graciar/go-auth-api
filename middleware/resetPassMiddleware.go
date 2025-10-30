package middleware

import (
	"go-auth/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResetTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		resetToken := c.Request.Header.Get("token")
		if resetToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no authorization header provided"})
			c.Abort()
			return
		}
		claims, err := helpers.ValidateToken(resetToken)

		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		if claims.TokenType != "reset" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Next()
	}
}
