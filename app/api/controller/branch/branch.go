package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	branchSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// BranchController 处理与分支相关的 HTTP 请求
type BranchController struct {
	branchSvc branchSvc.BranchInterface
}

// NewBranchController 创建分支处理器实例
func NewBranchController(branchService branchSvc.BranchInterface) *BranchController {
	return &BranchController{
		branchSvc: branchService,
	}
}

// @Summary 创建分支
// @Description 新增一个分支
// @Tags branches
// @Accept application/json
// @Produce application/json
// @Param request body branch.CreateBranchRequest true "创建分支请求参数"
// @Success 200 {object} resp.ObjectResponse{data=branch.BranchResponse} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /branches [post]
func (ctl *BranchController) Create(c *gin.Context) {

	log := logger.GetLogger(c.Request.Context())

	var reqData branch.CreateBranchRequest
	var err error
	// TODO: 获取操作人
	// 获取仓库
	repoRef, err := req.GetRepoRefFromPath(c)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchPathParamError)
		resp.BadRequest(c, err)
		return
	}
	log.Debugln(repoRef)

	// TODO: 检查是否为镜像仓库
	// TODO: 校验用户是否有仓库的 Push 权限

	// 解析请求体
	if err = c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// TODO: 如果未指定目标分支，使用仓库默认分支作为新分支的基准分支
	// - 判断是否允许绕过规则：如果绕过了保护规则，记录审计日志
	// - 获取保护规则的同时，判断当前用户是否为仓库所有者
	// - 根据保护规则验证分支名，以及用户的创建权限

	// 如果是 DryRunRules，只返回规则检查结果，不实际创建
	// 如果有致命违规且未绕过，直接返回错误

	// 调用 service 层创建分支
	respData, err := ctl.branchSvc.Create(c.Request.Context(), reqData)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	// 记录审计日志
	resp.Created(c, i18n.BranchCreateSuccess, respData)
}
