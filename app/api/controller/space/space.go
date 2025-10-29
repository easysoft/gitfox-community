package space

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	spaceSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/space"
	pkg_errors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
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

// @Summary 创建空间
// @Description 新增一个空间，需保证未删除空间中名称唯一，负责人默认为当前用户
// @Tags spaces
// @Accept application/json
// @Produce application/json
// @Param request body space.CreateSpaceRequest true "创建空间请求参数"
// @Success 200 {object} resp.ObjectResponse{data=space.SpaceResponse} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /spaces [post]
func (ctl *SpaceController) Create(c *gin.Context) {
	var req space.CreateSpaceRequest
	var err error

	// 解析请求体
	if err = c.ShouldBindJSON(&req); err != nil {
		err = pkg_errors.WithI18nKey(err, "space.create_param_error")
		resp.BadRequest(c, err)
		return
	}

	// 获取包含 logger 和其他上下文信息的 context 传递给 Service 层
	// 调用服务层创建空间
	spaceObj, err := ctl.spaceSvc.Create(logger.GetContextWithLogger(c), req)
	if err != nil {
		err = pkg_errors.WithI18nKey(err, "space.create_error")
		resp.InternalError(c, err)
		return
	}

	resp.Created(c, "space.create_success", spaceObj)
}

// @Summary 查询空间列表
// @Description 分页查询空间列表，支持按名称、权限、负责人筛选
// @Tags spaces
// @Accept application/json
// @Produce application/json
// @Param name query string false "空间名称（模糊查询）"
// @Param permission query string false "空间权限（open-私有，private-公开）" Enums(open, private)
// @Param owner query string false "负责人"
// @Param page query int false "页码（默认1）" default(1) minimum(1)
// @Param pageSize query int false "每页条数（默认20）" default(20) minimum(1) maximum(100)
// @Success 200 {object} resp.ListResponse{data=[]space.SpaceResponse} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /spaces [get]
func (ctl *SpaceController) List(c *gin.Context) {
	resp.ListSuccess(c, "space.list_success",
		resp.ListArgs{}, "0")
}

// @Summary 查询单个空间
// @Description 根据ID查询单个空间详情
// @Tags spaces
// @Accept application/json
// @Produce application/json
// @Param id path int true "空间ID" minimum(1)
// @Success 200 {object} resp.ObjectResponse{data=space.SpaceResponse} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "ID格式错误"
// @Failure 404 {object} resp.ObjectResponse "空间不存在或已删除"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /spaces/{id} [get]
func (ctl *SpaceController) Get(c *gin.Context) {
	resp.OK(c, "space.get_success", nil)
}

// @Summary 更新空间
// @Description 根据ID更新空间信息，未传入字段保持原值，名称更新需重新校验唯一性
// @Tags spaces
// @Accept application/json
// @Produce application/json
// @Param id path int true "空间ID" minimum(1)
// @Param request body space.UpdateSpaceRequest true "更新空间请求参数（字段可选）"
// @Success 200 {object} resp.ObjectResponse{data=space.SpaceResponse} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 404 {object} resp.ObjectResponse "空间不存在或已删除"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /spaces/{id} [put]
func (ctl *SpaceController) Update(c *gin.Context) {
	resp.OK(c, "space.update_success", nil)
}

// @Summary 删除空间
// @Description 逻辑删除空间，删除前需校验空间下无代码库和制品库
// @Tags spaces
// @Accept application/json
// @Produce application/json
// @Param id path int true "空间ID" minimum(1)
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "ID格式错误"
// @Failure 404 {object} resp.ObjectResponse "空间不存在或已删除"
// @Failure 500 {object} resp.ObjectResponse "删除失败（如存在关联资源）"
// @Router /spaces/{id} [delete]
func (ctl *SpaceController) Delete(c *gin.Context) {
	resp.OK(c, "space.delete_success", nil)
}

// @Summary 恢复已删除空间
// @Description 恢复已逻辑删除的空间
// @Tags spaces
// @Accept application/json
// @Produce application/json
// @Param request body space.UndeleteSpaceRequest true "恢复空间请求参数"
// @Success 200 {object} resp.ObjectResponse "恢复成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "空间不存在或未删除"
// @Failure 500 {object} resp.ObjectResponse "恢复失败"
// @Router /spaces/undelete [post]
func (ctl *SpaceController) Undelete(c *gin.Context) {
	resp.OK(c, "space.undelete_success", nil)
}
