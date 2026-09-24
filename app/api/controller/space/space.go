// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package space

import (
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/model/space"
	spaceSvc "github.com/easysoft/gitfox-community/app/service/space"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
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
	if err = reqData.Validate(); err != nil {
		resp.BadRequest(c, err)
		return
	}

	// 获取包含 logger 和其他上下文信息的 context 传递给 Service 层
	// 调用服务层创建空间
	spaceObj, err := ctl.spaceSvc.Create(c.Request.Context(), reqData)
	if err != nil {
		i18nKey := errors.GetI18nKey(err)
		if i18nKey == i18n.DbDuplicate {
			err = errors.WithI18nKey(err, i18n.SpaceCreateDuplicate)
			resp.BadRequest(c, err)
			return
		}
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
// @Param code query string false "空间标识（模糊查询）"
// @Param permission query string false "空间权限（open-私有，private-公开）" Enums(open, private)
// @Param owner query string false "负责人"
// @Param ids query string false "空间ID,逗号分隔"
// @Param deleted query bool false "是否包含已删除，默认false"
// @Param isOpen query bool false "是否公开空间，默认false"
// @Param page query int false "页码（默认1）" default(1) minimum(1)
// @Param pageSize query int false "每页条数（默认20）" default(20) minimum(1) maximum(2000)
// @Success 200 {object} resp.ListResponse{data=[]space.SpaceResponse,listArgs=resp.ListArgs}"查询成功：返回空间列表+分页元数据"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /spaces [get]
func (ctl *SpaceController) List(c *gin.Context) {
	var requestData space.ListSpaceFilter
	var err error
	if err = c.ShouldBindQuery(&requestData); err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	spaceObj, err := ctl.spaceSvc.List(c.Request.Context(), requestData)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpaceListError)
		resp.InternalError(c, err)
		return
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
	obj, err := ctl.spaceSvc.GetByID(c.Request.Context(), spaceID)
	if err != nil {
		i18nKey := errors.GetI18nKey(err)
		if i18nKey == i18n.DbNotfound {
			err = errors.WithI18nKey(err, i18n.SpaceNotFound)
			resp.NotFound(c, err)
			return
		}
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
	spaceObj, err := ctl.spaceSvc.Update(c.Request.Context(), objID, reqData)
	if err != nil {
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
	if err = ctl.spaceSvc.Delete(c.Request.Context(), objID); err != nil {
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
// @Param id path int true "空间ID" minimum(1)
// @Success 200 {object} resp.ObjectResponse "恢复成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "空间不存在或未删除"
// @Failure 500 {object} resp.ObjectResponse "恢复失败"
// @Router /spaces/{id}/restore [post]
func (ctl *SpaceController) Restore(c *gin.Context) {
	var err error
	// 解析 path id
	objID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(c, err)
		return
	}

	// 调用服务层恢复空间
	if err = ctl.spaceSvc.Restore(c.Request.Context(), objID); err != nil {
		// 判断错误类型，返回相应的HTTP状态码
		err = errors.WithI18nKey(err, i18n.SpaceRestoreError)
		resp.InternalError(c, err)
		return
	}
	resp.OK(c, i18n.SpaceRestoreSuccess, nil)
}

// @Summary 订阅空间事件
// @Description 订阅空间事件，实时获取空间内的操作事件
// @Tags spaces
// @Accept application/json
// @Produce application/json
// @Param id path int true "空间ID" minimum(1)
// @Success 200 {object} resp.ObjectResponse "事件流"
// @Failure 400 {object} resp.ObjectResponse "ID格式错误"
// @Failure 404 {object} resp.ObjectResponse "空间不存在或已删除"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /spaces/{id}/events [get]
func (ctl *SpaceController) Events(c *gin.Context) {
	var err error
	logger := logger.GetLogger(c.Request.Context())
	spaceID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(c, err)
		return
	}

	chEvents, chErr, sseCancel, err := ctl.spaceSvc.Events(c.Request.Context(), spaceID)
	if err != nil {
		resp.InternalError(c, err)
		return
	}
	defer func() {
		if err := sseCancel(c.Request.Context()); err != nil {
			logger.Errorf("failed to cancel sse stream for space '%d' ,err:%v", spaceID, err)
		}
	}()

	resp.StreamSSE(c.Request.Context(), c.Writer, c.Done(), chEvents, chErr)
}
