package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/scutrobotlab/rm-search/svc"
	"net/http"
)

// AdminAuthMiddleware 管理端权限中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		expected := "Bearer " + svc.Ctx().Config.AdminToken
		if token != expected {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
