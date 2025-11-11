package repo

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	urlPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	repoSvc     repo.Interface
	urlProvider urlPkg.Provider
}

// NewRepoController 创建代码库处理器实例
func NewRepoController(repoSvc repo.Interface, urlProvider urlPkg.Provider) *Controller {
	return &Controller{
		repoSvc:     repoSvc,
		urlProvider: urlProvider,
	}
}

// CreateRemoteRepo 创建远端仓库
// @Summary 创建远端仓库
// @Description 基于传入参数创建远端Git仓库，自动处理GitIgnore模板、License协议、README文件生成，支持派生仓库配置
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param body body repoModel.CreateRepoReq true "创建仓库请求参数"
// @Success 201 {object} resp.ObjectResponse{data=repoModel.CreateRepoResp} "创建成功（返回仓库核心信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如identifier格式非法、spaceID为空等）"
// @Error 404 {object} resp.ObjectResponse "空间不存在（传入的spaceID对应的空间未找到）"
// @Error 409 {object} resp.ObjectResponse "仓库标识冲突（同一空间内identifier已存在）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos [post]
func (ctl *Controller) Create(c *gin.Context) {
	reqData := new(repoModel.CreateRepoReq)
	if err := c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	err := reqData.Validate()
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	ctl.repoSvc.Create(logger.GetContextWithLogger(c), reqData)
}

func (ctl *Controller) GetContent(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params req.RepoId
	var err error

	_, ok := req.GetAuthSession(ctx)
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindUri(&params); err != nil || params.RepoID == 0 {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		resp.BadRequest(ctx, err)
		return
	}

	// gitRef 仓库引用，例如：refs/heads/main
	gitRef := req.QueryParamOrDefault(ctx, req.QueryParamGitRef, "refs/heads/main")

	// includeCommit 是否包含提交信息
	includeCommit, err := req.QueryParamAsBoolOrDefault(ctx, req.QueryParamIncludeCommit, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPullPermissionError)
		resp.BadRequest(ctx, err)
		return
	}

	// flattenDirectories 是否展平目录结构
	flattenDirectories, err := req.QueryParamAsBoolOrDefault(ctx, req.QueryParamFlattenDirectories, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 在像 /v1/repos/{repoRef}/contents/* 这样的路由下
	// * 表示仓库内的文件或目录的相对路径。
	// 例如请求 /v1/repos/12/contents/src/main.go ，这里的余量就是 src/main.go
	path := req.PathParamOrEmpty(ctx, req.PathParamRemainder)

	logger.Infof("origin url %s , repoID: %d,gitRef: %s, path: %s,includeCommit: %v,flattenDirectories: %v", ctx.Request.URL.String(), params.RepoID, gitRef, path, includeCommit, flattenDirectories)

	fmt.Println(ctx.Param("path"))
}
