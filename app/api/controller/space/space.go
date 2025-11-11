package space

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	spaceSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// SpaceController 处理与空间相关的 HTTP 请求
type SpaceController struct {
	spaceSvc spaceSvc.Interface
}

// NewSpaceController 创建空间处理器实例
func NewSpaceController(spaceService spaceSvc.Interface) *SpaceController {
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
	var reqData space.CreateSpaceRequest
	var err error

	// 解析请求体
	if err = c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// 获取包含 logger 和其他上下文信息的 context 传递给 Service 层
	// 调用服务层创建空间
	spaceObj, err := ctl.spaceSvc.Create(logger.GetContextWithLogger(c), reqData)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceCreateError)
		resp.InternalError(c, err)
		return
	}

	resp.Created(c, i18n.SpaceCreateSuccess, spaceObj)
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
	var req space.ListSpaceFilter
	var err error
	if err = c.ShouldBindQuery(&req); err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	spaceObj, err := ctl.spaceSvc.List(logger.GetContextWithLogger(c), req)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceListError)
		resp.InternalError(c, err)
	}
	resp.ListSuccess(c, spaceObj.Data,
		resp.ListArgs{
			Total:    spaceObj.Total,
			Page:     int64(spaceObj.Page),
			PageSize: int64(spaceObj.PageSize),
		}, i18n.SpaceListSuccess, nil)
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
func (ctl *SpaceController) GetByID(c *gin.Context) {
	spaceID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(c, err)
		return
	}

	obj, err := ctl.spaceSvc.GetByID(logger.GetContextWithLogger(c), spaceID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceQueryError)
		resp.InternalError(c, err)
		return
	}
	resp.OK(c, i18n.SpaceQuerySuccess, obj)
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
	var reqData space.UpdateSpaceRequest
	var err error

	// 解析 path id
	objID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(c, err)
		return
	}

	// 解析请求体
	if err = c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// 调用服务层更新空间
	spaceObj, err := ctl.spaceSvc.Update(logger.GetContextWithLogger(c), objID, reqData)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceUpdateError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.SpaceUpdateSuccess, spaceObj)
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
	var err error

	// 解析 path id
	objID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(c, err)
		return
	}

	// 调用服务层删除空间
	if err = ctl.spaceSvc.Delete(logger.GetContextWithLogger(c), objID); err != nil {
		// 判断错误类型，返回相应的HTTP状态码
		if errors.IsNotFound(err) {
			err = errors.WithI18nKey(err, "space.not_found")
			resp.NotFound(c, err)
			return
		}
		// 关联资源错误或其他业务错误
		err = errors.WithI18nKey(err, "space.delete_error")
		resp.InternalError(c, err)
		return
	}
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
	var reqData space.UndeleteSpaceRequest
	var err error

	// 解析请求体
	if err = c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// 调用服务层恢复空间
	if err = ctl.spaceSvc.Undelete(logger.GetContextWithLogger(c), reqData); err != nil {
		// 判断错误类型，返回相应的HTTP状态码
		if errors.IsNotFound(err) {
			err = errors.WithI18nKey(err, "space.not_found")
			resp.NotFound(c, err)
			return
		}
		// 如果是名称冲突或其他业务错误
		err = errors.WithI18nKey(err, "space.undelete_error")
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, "space.undelete_success", nil)
}
