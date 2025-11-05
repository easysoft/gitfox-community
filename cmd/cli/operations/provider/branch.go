package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	branchSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/branch"
)

var BranchWireSet = wire.NewSet(
	branch.NewBranchController,
	wire.Bind(new(branchSvc.BranchInterface), new(*branchSvc.BranchSvc)),
	branchSvc.NewBranchSvc,
)
