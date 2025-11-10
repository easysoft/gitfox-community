package repo

import (
	"fmt"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"github.com/gin-gonic/gin"
)

type RepoController struct {
	repoSvc repo.Interface
}

// NewRepoController 创建代码库处理器实例
func NewRepoController(repoSvc repo.Interface) *RepoController {
	return &RepoController{
		repoSvc: repoSvc,
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
// @Router /api/v1/repos [post]
func (ctl *RepoController) Create(c *gin.Context) {
	fmt.Println(repoModel.CreateRepoReq{})

}
