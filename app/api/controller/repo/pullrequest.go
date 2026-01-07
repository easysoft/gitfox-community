package repo

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Metadata 获取代码Pull Request 元数据
// @Summary 获取代码Pull Request 元数据
// @Description 根据传入的仓库 ID 和 Pull Request 编号获取 Pull Request 元数据
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pullreqNumber path uint true "Pull Request 编号"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.DiffStats} "获取成功（返回PR的差异统计信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如pullreqNumber格式非法、ID为空等）"
// @Error 404 {object} resp.ObjectResponse "代码库不存在（传入的repoID对应的空间未找到）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/pullreq/{pullreqNumber}/metadata [get]
func (ctl *Controller) Metadata(c *gin.Context) {
	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	repoID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}
	if repoID <= 0 {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}

	pullreqID, err := req.PathParamAsPositiveUint(c, req.PullreqNumber)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}
	if pullreqID <= 0 {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}

	data, err := ctl.repoSvc.GetMetadata(c.Request.Context(), session, repoID, pullreqID)
	if err != nil {
		resp.BadRequest(c, err)
		return
	}

	resp.OK(c, i18n.PullrequestQuerySuccess, data)
}

// GetDiffStats 获取代码Pull Request 差异统计信息
// @Summary 获取代码Pull Request 差异统计信息
// @Description 根据传入的仓库 ID 和 路径获取 Pull Request 差异统计信息
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pathToPath path string true "路径"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.DiffStats} "获取成功（返回差异统计信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如pathToPath格式非法、为空等）"
// @Error 404 {object} resp.ObjectResponse "代码库不存在（传入的repoID对应的空间未找到）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/diff-stats/{pathToPath} [get]
func (ctl *Controller) GetDiffStats(c *gin.Context) {
	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	repoID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}
	if repoID <= 0 {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}

	pathToPath, err := req.PathParamOrError(c, req.PathToPath)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}
	if pathToPath == "" {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}

	pathToPath = strings.TrimPrefix(pathToPath, "/")

	data, err := ctl.repoSvc.GetDiffStats(c.Request.Context(), session, repoID, pathToPath)
	if err != nil {
		resp.BadRequest(c, err)
		return
	}

	resp.OK(c, i18n.PullrequestQuerySuccess, data)
}

// MergeCheck 获取代码Pull Request 合并检查信息
// @Summary 获取代码Pull Request 合并检查信息
// @Description 根据传入的仓库 ID 和 路径获取 Pull Request 合并检查信息
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pathToPath path string true "路径"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.MergeCheck} "获取成功（返回合并检查信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如pathToPath格式非法、为空等）"
// @Error 404 {object} resp.ObjectResponse "代码库不存在（传入的repoID对应的空间未找到）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/merge-check/{pathToPath} [post]
func (ctl *Controller) MergeCheck(c *gin.Context) {
	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	repoID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}
	if repoID <= 0 {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}

	pathToPath, err := req.PathParamOrError(c, req.PathToPath)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}
	if pathToPath == "" {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(c, err)
		return
	}

	pathToPath = strings.TrimPrefix(pathToPath, "/")

	data, err := ctl.repoSvc.MergeCheck(c.Request.Context(), session, repoID, pathToPath)
	if err != nil {
		resp.BadRequest(c, err)
		return
	}

	resp.OK(c, i18n.PullrequestMergeCheckSuccess, data)
}
