package router

import (
	"github.com/gin-gonic/gin"
)

// registerSvnRoutes 注册 SVN 相关路由
func registerSvnRoutes(r *gin.RouterGroup, controller *Controller) {
	svn := r.Group("/repos/:id/svn")
	{
		svn.POST("/exec", controller.Svn.Exec) // 统一执行入口
	}
}
