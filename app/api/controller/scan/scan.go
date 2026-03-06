package scan

import scanService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"

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
