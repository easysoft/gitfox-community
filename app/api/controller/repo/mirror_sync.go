package repo

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// swagger 类型解析需要用到 repoModel，通过包级变量引用避免编译器报 unused import。
var _ = repoModel.Progress{}

// MirrorSync 手动触发镜像仓库同步
// @Summary 手动触发镜像仓库同步
// @Description 对指定镜像仓库手动触发一次同步任务
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param repoID query int64 true "镜像仓库 ID"
// @Success 200 {object} resp.ObjectResponse "同步任务已触发"
// @Error 400 {object} resp.ObjectResponse "参数错误（如 repoID 格式非法）"
// @Error 401 {object} resp.ObjectResponse "未授权"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如任务调度失败）"
// @Router /repos/mirror-sync [post]
func (ctl *Controller) MirrorSync(c *gin.Context) {
	repoID, err := req.QueryParamAsPositiveInt64OrError(c, "repoID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryParamError)
		resp.BadRequest(c, err)
		return
	}

	if _, ok := req.GetAuthSession(c.Request.Context()); !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if err := ctl.importerSvc.CreateMirrorSyncJob(c.Request.Context(), repoID); err != nil {
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RepoMirrorSyncSuccess, nil)
}

// MirrorSyncProgress 查询镜像同步进度
// @Summary 查询镜像同步进度
// @Description 查询指定镜像仓库的同步任务进度
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param repoID query int64 true "镜像仓库 ID"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.Progress} "查询成功"
// @Error 400 {object} resp.ObjectResponse "参数错误（如 repoID 格式非法）"
// @Error 401 {object} resp.ObjectResponse "未授权"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /repos/mirror-sync-progress [get]
func (ctl *Controller) MirrorSyncProgress(c *gin.Context) {
	repoID, err := req.QueryParamAsPositiveInt64OrError(c, "repoID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryParamError)
		resp.BadRequest(c, err)
		return
	}

	if _, ok := req.GetAuthSession(c.Request.Context()); !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	progress, err := ctl.importerSvc.GetMirrorSyncProgress(c.Request.Context(), repoID)
	if err != nil {
		resp.BadRequest(c, err)
		return
	}

	resp.OK(c, i18n.RepoMirrorSyncProgressSuccess, progress)
}
