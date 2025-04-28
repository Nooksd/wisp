package middleware

import (
	"net/http"
	"strings"

	"wisp/src/service"

	"github.com/gin-gonic/gin"
)

func JWTAuth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		var tokenStr string

		if h != "" || strings.HasPrefix(h, "Bearer ") {
			tokenStr = strings.TrimPrefix(h, "Bearer ")
		} else {
			if tokenStr = c.Query("token"); tokenStr == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token ausente"})
				return
			}
		}

		claims, err := authSvc.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("isAdmin", claims.IsAdmin)
		c.Set("sid", claims.ID)
		c.Next()
	}
}
