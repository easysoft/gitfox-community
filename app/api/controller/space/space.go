package space

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	spaceSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// SpaceController 处理与空间相关的 HTTP 请求
type SpaceController struct {
	spaceSvc spaceSvc.SpaceService
}

// NewSpaceController 创建空间处理器实例
func NewSpaceController(spaceService spaceSvc.SpaceService) *SpaceController {
	return &SpaceController{
		spaceSvc: spaceService,
	}
}

// Create 处理创建空间请求
// @Summary      创建空间
// @Description  创建新的空间
// @Tags         空间管理
// @Accept       json
// @Produce      json
// @Param        request   body      space.CreateSpaceRequest       true  "创建空间请求"
// @Success      201       {object}  space.CreateSpaceResponse             "创建成功"
// @Failure      400       {object}  map[string]string       "请求参数错误"
// @Failure      500       {object}  map[string]string       "服务器内部错误"
// @Router       /spaces [post]
// @Security     ApiKeyAuth
func (h *SpaceController) Create(c *gin.Context) {
	// 从 gin.Context 中获取带有 trace_id 的 logger
	log := logger.GetLoggerFromGin(c)
	log.Debugf("[Controller] 开始处理创建空间请求")

	var req space.CreateSpaceRequest

	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("[Controller] 参数解析失败: %v", err)
		resp.ObjectBadRequest(c, "space.create_param_error", nil)
		return
	}

	log.Debugf("[Controller] 参数解析成功，准备调用 Service 层, space_name=%s", req.Name)

	// 获取包含 logger 和其他上下文信息的 context 传递给 Service 层
	// 调用服务层创建空间
	spaceObj, err := h.spaceSvc.Create(logger.GetContextWithLogger(c), req)
	if err != nil {
		log.Errorf("[Controller] 创建空间失败: %v", err)
		resp.ObjectInternalError(c, "space.create_error", nil)
		return
	}

	log.Debugf("[Controller] 空间创建成功, space_id=%s", spaceObj.ID)
	resp.ObjectCreated(c, "space.create_success", spaceObj)
}
