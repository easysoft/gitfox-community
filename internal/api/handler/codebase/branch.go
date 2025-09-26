package codebase

import (
	"net/http"

	"github.com/easysoft/gitfox-pms/internal/api/middleware"
	"github.com/easysoft/gitfox-pms/internal/constants"
	"github.com/easysoft/gitfox-pms/internal/service/codebase/branch"
	"github.com/easysoft/gitfox-pms/pkg/logger"

	"github.com/gin-gonic/gin"
)

// BranchHandler 分支API处理器
type BranchHandler struct {
	branchService branch.BranchService
	logger        logger.Logger
}

// NewBranchHandler 创建分支处理器实例
func NewBranchHandler(branchService branch.BranchService) *BranchHandler {
	return &BranchHandler{
		branchService: branchService,
		// logger:        logger,
	}
}

// Create 处理创建分支请求
func (h *BranchHandler) Create(c *gin.Context) {
	var req branch.CreateBranchRequest

	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// 从URL获取代码库ID
	req.RepoID = c.Param("repo_id")
	// 从上下文获取用户ID(由认证中间件设置)
	req.UserID = c.GetString(middleware.ContextUserID)

	// 调用服务层创建分支
	branch, err := h.branchService.Create(c.Request.Context(), req)
	if err != nil {
		logger.Error("创建分支失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, branch)
}

// Get 处理获取分支详情请求
func (h *BranchHandler) Get(c *gin.Context) {
	branchID := c.Param("branch_id")

	branch, err := h.branchService.Get(c.Request.Context(), branchID)
	if err != nil {
		logger.Error("获取分支失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if branch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrBranchNotFound})
		return
	}

	c.JSON(http.StatusOK, branch)
}
