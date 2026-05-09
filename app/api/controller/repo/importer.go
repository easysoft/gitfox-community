package repo

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ImportRemoteRepo 导入远端仓库
// @Summary 导入远端仓库
// @Description 基于传入参数导入远端Git仓库，自动处理GitIgnore模板、License协议、README文件生成，支持派生仓库配置
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param body body repoModel.ImportInput true "导入仓库请求参数"
// @Success 201 {object} resp.ObjectResponse{data=repoModel.RepositoryOutput} "导入成功（返回仓库核心信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如identifier格式非法、spaceID为空等）"
// @Error 404 {object} resp.ObjectResponse "空间不存在（传入的spaceID对应的空间未找到）"
// @Error 409 {object} resp.ObjectResponse "仓库标识冲突（同一空间内identifier已存在）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/import [post]
func (ctl *Controller) Import(c *gin.Context) {
	in := new(repoModel.ImportInput)
	if err := c.ShouldBindJSON(&in); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if session.User.Email == "" {
		resp.BadRequest(c, errors.WithI18nKey(nil, i18n.RepoEmailRequired))
		return
	}

	remoteRepository, provider, err := ctl.importerSvc.LoadRepositoryFromProvider(c.Request.Context(), in.Provider, in.ProviderRepo)
	if err != nil {
		resp.BadRequest(c, err)
		return
	}

	repoData, isPublic := remoteRepository.ToRepo(
		in.SpaceID,
		in.Identifier,
		in.Description,
		session.User,
	)

	// 创建代码库
	repo, err := ctl.importerSvc.Create(c.Request.Context(), session, repoData)
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

	if err := ctl.importerSvc.Run(c.Request.Context(),
		provider,
		repo,
		isPublic,
		remoteRepository.CloneURL,
		false,
	); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoImportError)
		resp.BadRequest(c, err)
		return
	}

	resp.Created(c, i18n.RepoImportSuccess, repoModel.RepositoryOutput{
		Repo:      *repo,
		IsPublic:  isPublic,
		Importing: repo.Status != int(constant.RepoStateActive),
	})
}
