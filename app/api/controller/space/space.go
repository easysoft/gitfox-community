package space

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/space"
	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
)

// SpaceController 处理与空间相关的 HTTP 请求
type SpaceController struct {
	spaceSvc space.SpaceService
}

// NewSpaceController 创建空间处理器实例
func NewSpaceController(spaceService space.SpaceService) *SpaceController {
	return &SpaceController{
		spaceSvc: spaceService,
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
func (h *SpaceController) Create(c *gin.Context) {
	var req spaceModel.CreateSpaceRequest

	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		// h.logger.Error("参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}


	// 调用服务层创建分支
	spaceObj, err := h.spaceSvc.Create(c.Request.Context(), req)
	if err != nil {
		//h.logger.Error("创建分支失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, spaceObj)
}