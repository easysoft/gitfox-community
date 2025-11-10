package repo

import (
	"fmt"
	"net/url"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	repoSvc repo.Interface
	logger  *logger.SugaredMateLogger
}

// NewRepoController 创建代码库处理器实例
func NewRepoController(repoSvc repo.Interface, logger *logger.SugaredMateLogger) *RepoController {
	return &RepoController{
		repoSvc: repoSvc,
		logger:  logger,
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

func (ctl *RepoController) GetContent(c *gin.Context) {
	var params req.RepoURI
	var err error

	_, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	if err = c.ShouldBindUri(&params); err != nil || params.RepoRef == "" {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		resp.BadRequest(c, err)
		return
	}

	val, err := url.PathUnescape(params.RepoRef)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		resp.BadRequest(c, err)
		return
	}

	// includeCommit 是否包含提交信息
	includeCommit, err := req.QueryParamAsBoolOrDefault(c, req.QueryParamIncludeCommit, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPullPermissionError)
		resp.BadRequest(c, err)
		return
	}
	_ = includeCommit

	// flattenDirectories 是否展平目录结构
	flattenDirectories, err := req.QueryParamAsBoolOrDefault(c, req.QueryParamFlattenDirectories, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(c, err)
		return
	}
	_ = flattenDirectories
	ctl.logger.Infof("origin url %s , repoRef: %s", c.Request.URL.String(), val)
	fmt.Println(c.Param("path"))
}

func (ctl *RepoController) GetContent(c *gin.Context) {
	var params req.RepoURI
	var err error

	_, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	if err = c.ShouldBindUri(&params); err != nil || params.RepoRef == "" {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		resp.BadRequest(c, err)
		return
	}

	val, err := url.PathUnescape(params.RepoRef)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		resp.BadRequest(c, err)
		return
	}

	// includeCommit 是否包含提交信息
	includeCommit, err := req.QueryParamAsBoolOrDefault(c, req.QueryParamIncludeCommit, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPullPermissionError)
		resp.BadRequest(c, err)
		return
	}
	_ = includeCommit

	// flattenDirectories 是否展平目录结构
	flattenDirectories, err := req.QueryParamAsBoolOrDefault(c, req.QueryParamFlattenDirectories, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(c, err)
		return
	}
	_ = flattenDirectories
	ctl.logger.Infof("origin url %s , repoRef: %s", c.Request.URL.String(), val)
	fmt.Println(c.Param("path"))
}
