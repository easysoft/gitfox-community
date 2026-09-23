// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	repoService "github.com/easysoft/gitfox-community/app/service/repo"
	scanService "github.com/easysoft/gitfox-community/app/service/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

type ScanController struct {
	ruleSetSvc       scanService.IRuleSet
	ruleSvc          scanService.IRule
	solutionSvc      scanService.ISolution
	planSvc          scanService.IScanPlan
	triggerSvc       scanService.IScanTrigger
	conditionSvc     scanService.IScanCondition
	planConditionSvc scanService.IScanPlanCondition
	taskSvc          scanService.IScanTask
	repoSvc          repoService.Interface
	gitPkg           git.Interface
	aicodereviewSvc  scanService.IAIReview
}

// NewScanController 创建扫描处理器实例
func NewScanController(
	ruleSetSvc scanService.IRuleSet,
	ruleSvc scanService.IRule,
	solutionSvc scanService.ISolution,
	taskSvc scanService.IScanTask,
	planSvc scanService.IScanPlan,
	triggerSvc scanService.IScanTrigger,
	conditionSvc scanService.IScanCondition,
	planConditionSvc scanService.IScanPlanCondition,
	repoSvc repoService.Interface,
	gitPkg git.Interface,
	aicodereviewSvc scanService.IAIReview,
) *ScanController {
	return &ScanController{
		ruleSetSvc:       ruleSetSvc,
		ruleSvc:          ruleSvc,
		solutionSvc:      solutionSvc,
		planSvc:          planSvc,
		triggerSvc:       triggerSvc,
		conditionSvc:     conditionSvc,
		planConditionSvc: planConditionSvc,
		repoSvc:          repoSvc,
		taskSvc:          taskSvc,
		gitPkg:           gitPkg,
		aicodereviewSvc:  aicodereviewSvc,
	}
}

func (c *ScanController) pathRepoID(ctx *gin.Context, errKey i18n.I18nKey) (uint, bool) {
	repoID, err := req.PathParamAsNaturalNum(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, errKey))
		return 0, false
	}
	return repoID, true
}

func (c *ScanController) pathRepoAndPlanID(ctx *gin.Context, errKey i18n.I18nKey) (uint, uint, bool) {
	repoID, ok := c.pathRepoID(ctx, errKey)
	if !ok {
		return 0, 0, false
	}
	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, errKey))
		return 0, 0, false
	}
	return repoID, planID, true
}

func (c *ScanController) pathRepoPlanAndTriggerID(ctx *gin.Context, errKey i18n.I18nKey) (uint, uint, uint, bool) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, errKey)
	if !ok {
		return 0, 0, 0, false
	}
	triggerID, err := req.PathParamAsPositiveUint(ctx, "triggerID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, errKey))
		return 0, 0, 0, false
	}
	return repoID, planID, triggerID, true
}

func (c *ScanController) pathRepoPlanAndConditionID(ctx *gin.Context, errKey i18n.I18nKey) (uint, uint, uint, bool) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, errKey)
	if !ok {
		return 0, 0, 0, false
	}
	conditionID, err := req.PathParamAsPositiveUint(ctx, "conditionID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, errKey))
		return 0, 0, 0, false
	}
	return repoID, planID, conditionID, true
}
