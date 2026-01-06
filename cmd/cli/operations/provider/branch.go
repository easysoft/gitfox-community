package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/branch"
	branchSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/branch"
	branchStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/branch"
)

var BranchWireSet = wire.NewSet(
	branch.NewController,
	wire.Bind(new(branchSvc.Interface), new(*branchSvc.Service)),
	branchSvc.NewService,
	wire.Bind(new(branchStore.Interface), new(*branchStore.Store)),
	branchStore.NewBranchStore,
)
