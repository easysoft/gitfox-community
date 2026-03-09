package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

type ScanController struct {
	ruleSetSvc   scanService.IRuleSet
	ruleSvc      scanService.IRule
	solutionSvc  scanService.ISolution
	planSvc      scanService.IScanPlan
	triggerSvc   scanService.IScanTrigger
	conditionSvc scanService.IScanCondition
}

// NewScanController 创建扫描处理器实例
func NewScanController(
	ruleSetSvc scanService.IRuleSet,
	ruleSvc scanService.IRule,
	solutionSvc scanService.ISolution,
	planSvc scanService.IScanPlan,
	triggerSvc scanService.IScanTrigger,
	conditionSvc scanService.IScanCondition,
) *ScanController {
	return &ScanController{
		ruleSetSvc:   ruleSetSvc,
		ruleSvc:      ruleSvc,
		solutionSvc:  solutionSvc,
		planSvc:      planSvc,
		triggerSvc:   triggerSvc,
		conditionSvc: conditionSvc,
	}
}

func (c *ScanController) pathRepoID(ctx *gin.Context, errKey i18n.I18nKey) (uint, bool) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
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
