package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Key")
		if key == "" || key != os.Getenv("API_KEY") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}
		c.Next()
	}
}
