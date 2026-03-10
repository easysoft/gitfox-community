package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/scan"
	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	scanSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
)

func ProvideScanCronScheduler(scheduler *jobsSvc.Scheduler) scanSvc.CronScheduler {
	return scheduler
}

var ScanRuleSetWireSet = wire.NewSet(
	scan.NewScanController,
	ProvideScanCronScheduler,
	wire.Bind(new(scanSvc.IRuleSet), new(*scanSvc.RuleSetService)),
	scanSvc.NewRuleSetService,
	wire.Bind(new(scanSvc.IRule), new(*scanSvc.RuleService)),
	scanSvc.NewRuleService,
	wire.Bind(new(scanStore.IRuleSet), new(*scanStore.RuleSetStore)),
	scanStore.NewRuleSetStore,
	wire.Bind(new(scanSvc.ISolution), new(*scanSvc.SolutionService)),
	scanSvc.NewSolutionService,
	wire.Bind(new(scanStore.ISolution), new(*scanStore.SolutionStore)),
	scanStore.NewSolutionStore,
	wire.Bind(new(scanSvc.IScanPlan), new(*scanSvc.ScanPlanService)),
	scanSvc.NewScanPlanService,
	wire.Bind(new(scanStore.IScanPlan), new(*scanStore.ScanPlanStore)),
	scanStore.NewScanPlanStore,
	wire.Bind(new(scanSvc.IScanTrigger), new(*scanSvc.ScanTriggerService)),
	scanSvc.NewScanTriggerService,
	wire.Bind(new(scanStore.IScanTrigger), new(*scanStore.ScanTriggerStore)),
	scanStore.NewScanTriggerStore,
	wire.Bind(new(scanSvc.IScanCondition), new(*scanSvc.ScanConditionService)),
	scanSvc.NewScanConditionService,
	wire.Bind(new(scanStore.IScanCondition), new(*scanStore.ScanConditionStore)),
	scanStore.NewScanConditionStore,
)

var ScanRuleWireSet = wire.NewSet(
	wire.Bind(new(scanStore.IRule), new(*scanStore.RuleStore)),
	scanStore.NewRuleStore,
)
