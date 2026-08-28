package middleware

import (
	"petunia/internal/config"
	response "petunia/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
				token = "Bearer " + cookie
			}
		}
		if token == "" {
			response.Unauthorized(c, "missing token")
			c.Abort()
			return
		}

		parts := strings.SplitN(token, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Authorization not valid (expected: Bearer <token>)")
			c.Abort()
			return
		}

		claims, err := config.ParseTokenOfType(parts[1], config.AccessTokenType)
		if err != nil {
			response.Unauthorized(c, "token not valid or expired")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Next()
	}
}
