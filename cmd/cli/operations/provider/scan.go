package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/scan"
	scanSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
)

var ScanRuleSetWireSet = wire.NewSet(
	scan.NewRuleSetController,
	wire.Bind(new(scanSvc.IRuleSet), new(*scanSvc.RuleSetService)),
	scanSvc.NewRuleSetService,
	wire.Bind(new(scanStore.IRuleSet), new(*scanStore.RuleSetStore)),
	scanStore.NewRuleSetStore,
)

var ScanRuleWireSet = wire.NewSet(
	wire.Bind(new(scanStore.IRule), new(*scanStore.RuleStore)),
	scanStore.NewRuleStore,
)
