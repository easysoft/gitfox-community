package repo

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Import 导入代码库
// @Summary 导入代码库
// @Description 基于传入参数导入远端Git仓库
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param body body repoModel.ImportInput true "导入仓库请求参数"
// @Success 201 {object} resp.ObjectResponse{data=repoModel.RepositoryOutput} "导入成功（返回仓库核心信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如name格式非法、spaceID为空等）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/import [post]
func (ctl *Controller) Import(c *gin.Context) {
	in := new(repoModel.ImportInput)
	if err := c.ShouldBindJSON(&in); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	if err := in.Validate(); err != nil {
		resp.BadRequest(c, err)
		return
	}

	session, ok := req.GetAuthSessionWithEmail(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.RepoEmailRequired))
		return
	}

	remoteRepository, provider, err := ctl.importerSvc.LoadRepositoryFromProvider(
		c.Request.Context(),
		in.Provider,
		in.Provider.Slug,
	)
	if err != nil {
		resp.BadRequest(c, err)
		return
	}

	localRepoData, _ := remoteRepository.ConvertToRepo(
		in.SpaceID,
		in.Name,
		in.Desc,
		in.ACL,
		in.Mirror,
		in.Product,
		session.User,
		&provider,
	)

	// 创建代码库
	repo, err := ctl.importerSvc.Create(c.Request.Context(), session, localRepoData)
	if err != nil {
		i18nKey := errors.GetI18nKey(err)
		if i18nKey == i18n.DbDuplicate {
			err = errors.WithI18nKey(err, i18n.RepoCreateDuplicate)
			resp.BadRequest(c, err)
			return
		}
		resp.InternalError(c, err)
		return
	}

	noticeLang := i18n.GetLocaleFromGin(c)
	if err := ctl.importerSvc.Run(c.Request.Context(),
		provider,
		repo,
		remoteRepository.CloneURL,
		in.Mirror,
		noticeLang,
	); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoImportError)
		resp.BadRequest(c, err)
		return
	}

	resp.Created(c, i18n.RepoImportSuccess, repoModel.RepositoryOutput{
		Repo: *repo,
	})
}

// ImportProgress 查询导入进度
// @Summary 查询导入进度
// @Description 查询导入进度（仅适用于正在导入中的仓库）
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param repoID query uint true "导入仓库ID"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.Progress} "查询成功"
// @Error 400 {object} resp.ObjectResponse "参数错误（如repoID格式非法）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如数据库异常等）"
// @Router /repos/import-progress [get]
func (ctl *Controller) ImportProgress(c *gin.Context) {
	repoID, err := req.QueryParamAsPositiveInt64OrError(c, "repoID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryParamError)
		resp.BadRequest(c, err)
		return
	}

	progress, err := ctl.importerSvc.GetProgress(c.Request.Context(), uint(repoID))
	if err != nil {
		resp.BadRequest(c, err)
		return
	}

	resp.OK(c, i18n.RepoQueryImportProgressSuccess, progress)
}
