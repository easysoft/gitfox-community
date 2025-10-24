package codebase

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/middleware"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constants"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/codebase/branch"
)

// BranchHandler 分支API处理器
type BranchHandler struct {
	branchService branch.BranchService
}

// NewBranchHandler 创建分支处理器实例
func NewBranchHandler(branchService branch.BranchService) *BranchHandler {
	return &BranchHandler{
		branchService: branchService,
	}
}

// Create 处理创建分支请求
// @Summary      创建分支
// @Description  在指定代码库中创建新分支
// @Tags         分支管理
// @Accept       json
// @Produce      json
// @Param        repo_id   path      string  true  "代码库ID"
// @Param        request   body      branch.CreateBranchRequest true  "创建分支请求"
// @Success      201       {object}  branch.CreateBranchResponse "创建成功"
// @Failure      400       {object}  map[string]string          "请求参数错误"
// @Failure      500       {object}  map[string]string          "服务器内部错误"
// @Router       /repos/{repo_id}/branches [post]
// @Security     ApiKeyAuth
func (h *BranchHandler) Create(c *gin.Context) {
	var req branch.CreateBranchRequest

	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		// h.logger.Error("参数解析失败: %v", err)
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
		//h.logger.Error("创建分支失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, branch)
}

// Get 处理获取分支详情请求
// @Summary      获取分支详情
// @Description  根据分支ID获取分支的详细信息
// @Tags         分支管理
// @Accept       json
// @Produce      json
// @Param        branch_id path      string  true  "分支ID"
// @Success      200       {object}  branch.CreateBranchResponse "获取成功"
// @Failure      404       {object}  map[string]string           "分支不存在"
// @Failure      500       {object}  map[string]string           "服务器内部错误"
// @Router       /branches/{branch_id} [get]
// @Security     ApiKeyAuth
func (h *BranchHandler) Get(c *gin.Context) {
	branchID := c.Param("branch_id")

	branch, err := h.branchService.Get(c.Request.Context(), branchID)
	if err != nil {
		//h.logger.Error("获取分支失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if branch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrBranchNotFound})
		return
	}

	c.JSON(http.StatusOK, branch)
}
