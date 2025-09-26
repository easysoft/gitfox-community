package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"gitfox.zcorp.cc/git/demo/gitfox-pms/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// ContextUserID 上下文用户ID键名
const ContextUserID = "user_id"

// Auth 认证中间件，验证用户Token
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息"})
			c.Abort()
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误，应为Bearer <token>"})
			c.Abort()
			return
		}

		// 验证Token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token: " + err.Error()})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set(ContextUserID, claims.UserID)
		c.Next()
	}
}

// ResourceAuth 资源权限校验中间件
func ResourceAuth(resourceType, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := c.Get(ContextUserID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}
		fmt.Println(userID)

		// 获取资源ID
		resourceID := c.Param("repo_id")
		if resourceID == "" {
			resourceID = c.Param("branch_id")
		}

		// 检查权限
		// permService := GetPermissionService()
		// err := permService.CheckPermission(
		// 	c.Request.Context(),
		// 	userID.(string),
		// 	resourceType,
		// 	resourceID,
		// 	action,
		// )
		// if err != nil {
		// 	c.JSON(http.StatusForbidden, gin.H{"error": "权限不足: " + err.Error()})
		// 	c.Abort()
		// 	return
		// }

		c.Next()
	}
}
