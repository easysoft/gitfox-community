package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// CreateBranchType 创建分支类型
// @Summary 创建分支类型
// @Description 为指定仓库创建一个新的分支类型
// @Tags branch-types
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body branchModel.ReqForCreateBranchType true "创建分支类型请求参数"
// @Success 201 {object} resp.ObjectResponse{data=branchModel.BranchType} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 409 {object} resp.ObjectResponse "分支类型 Key 已存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/branch-types [post]
func (c *Controller) CreateBranchType(ctx *gin.Context) {
	// 解析 path id
	repoID, err := req.PathParamAsNaturalNum(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 获取会话
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if repoID != 0 {
		// 通过仓库 ID 获取仓库
		repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
		if err != nil || repoOut == nil {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
			resp.InternalError(ctx, err)
			return
		}
	}

	// 解析请求题
	var reqBodyArray []branchModel.ReqForCreateBranchType
	if err := ctx.ShouldBindJSON(&reqBodyArray); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	for _, v := range reqBodyArray {
		// 调用Check方法进行参数校验
		if err := v.Check(); err != nil {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
			resp.BadRequest(ctx, err)
			return
		}
	}

	// 如果是空数组
	if len(reqBodyArray) == 0 {
		err = errors.WithI18nKey(errors.New("request body cannot be empty"), i18n.BranchTypeBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// repoID=0 时不允许批量创建
	if repoID == 0 && len(reqBodyArray) > 1 {
		err = errors.WithI18nKey(errors.New("batch creation not allowed for system-level branch types"), i18n.BranchTypeBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 批量创建分支类型
	respData, err := c.branchSvc.BatchCreateBranchType(ctx.Request.Context(), reqBodyArray, session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.Created(ctx, i18n.BranchTypeCreateSuccess, respData)
}

// UpdateBranchType 更新分支类型
// @Summary 更新分支类型
// @Description 更新指定的分支类型信息
// @Tags branch-types
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param typeid path uint true "分支类型 ID"
// @Param request body branchModel.ReqForUpdateBranchType true "更新分支类型请求参数"
// @Success 200 {object} resp.ObjectResponse{data=branchModel.BranchType} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "分支类型不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/branch-types/{typeid} [put]
func (c *Controller) UpdateBranchType(ctx *gin.Context) {
	var reqBody branchModel.ReqForUpdateBranchType

	// 解析 path id
	repoID, err := req.PathParamAsNaturalNum(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	typeID, err := req.PathParamAsPositiveUint(ctx, "typeid")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 调用Check方法进行参数校验
	if err := reqBody.Check(); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 获取会话
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if repoID != 0 {
		// 通过仓库 ID 获取仓库
		repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
		if err != nil || repoOut == nil {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
			resp.InternalError(ctx, err)
			return
		}
	}

	// 调用 service 层更新分支类型
	respData, err := c.branchSvc.UpdateBranchType(ctx.Request.Context(), typeID, reqBody, session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.BranchTypeUpdateSuccess, respData)
}

// GetBranchType 获取分支类型详情
// @Summary 获取分支类型详情
// @Description 获取指定分支类型的详细信息
// @Tags branch-types
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param typeid path uint true "分支类型 ID"
// @Success 200 {object} resp.ObjectResponse{data=branchModel.BranchType} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "分支类型不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/branch-types/{typeid} [get]
func (c *Controller) GetBranchType(ctx *gin.Context) {
	// 解析 path id
	repoID, err := req.PathParamAsNaturalNum(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	typeID, err := req.PathParamAsPositiveUint(ctx, "typeid")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 获取会话
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if repoID != 0 {
		// 通过仓库 ID 获取仓库
		repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
		if err != nil || repoOut == nil {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
			resp.InternalError(ctx, err)
			return
		}
	}

	// 调用 service 层获取分支类型
	respData, err := c.branchSvc.GetBranchType(ctx.Request.Context(), typeID, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.BranchTypeQuerySuccess, respData)
}

// DeleteBranchType 删除分支类型
// @Summary 删除分支类型
// @Description 删除指定的分支类型
// @Tags branch-types
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param typeid path uint true "分支类型 ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "分支类型不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/branch-types/{typeid} [delete]
func (c *Controller) DeleteBranchType(ctx *gin.Context) {
	// 解析 path id
	repoID, err := req.PathParamAsNaturalNum(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	typeID, err := req.PathParamAsPositiveUint(ctx, "typeid")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 获取会话
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if repoID != 0 {
		// 通过仓库 ID 获取仓库
		repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
		if err != nil || repoOut == nil {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
			resp.InternalError(ctx, err)
			return
		}
	}

	// 调用 service 层删除分支类型
	err = c.branchSvc.DeleteBranchType(ctx.Request.Context(), typeID, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.BranchTypeDeleteSuccess, nil)
}

// ListBranchTypes 查询分支类型列表
// @Summary 查询分支类型列表
// @Description 分页查询仓库的分支类型列表
// @Tags branch-types
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body branchModel.ReqFilterBranchType true "查询分支类型请求参数"
// @Success 200 {object} resp.ListResponse{data=[]branchModel.BranchType,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/{id}/branch-types/list [post]
func (c *Controller) ListBranchTypes(ctx *gin.Context) {
	var reqFilter branchModel.ReqFilterBranchType

	// 解析 path id
	repoID, err := req.PathParamAsNaturalNum(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err := ctx.ShouldBindJSON(&reqFilter); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 设置默认值
	req.SetDefaults(&reqFilter)

	// 获取会话
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if repoID != 0 {
		// 通过仓库 ID 获取仓库
		repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
		if err != nil || repoOut == nil {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyParamError)
			resp.InternalError(ctx, err)
			return
		}
	}

	// 调用 service 层查询分支类型列表
	pageResult, err := c.branchSvc.ListBranchTypes(ctx.Request.Context(), reqFilter, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, pageResult.Data,
		resp.ListArgs{
			Total:    pageResult.Total,
			Page:     int64(pageResult.Page),
			PageSize: int64(pageResult.PageSize),
		}, i18n.BranchTypeQuerySuccess, nil)
}
