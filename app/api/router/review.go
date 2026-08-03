package router

import "github.com/gin-gonic/gin"

// registerReviewRoutes 注册评审相关路由
func registerReviewRoutes(r *gin.RouterGroup, controller *Controller) {
	rule := r.Group("/review")
	rule.POST("/:repoID/:ppmID/report", controller.Review.ReviewTaskUploadReport)
}
