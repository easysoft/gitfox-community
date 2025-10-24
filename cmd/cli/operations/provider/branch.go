package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/handler/codebase"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/codebase/branch"
	branchd "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/codebase/branch"
)

var BranchWireSet = wire.NewSet(
	codebase.NewBranchHandler,
	wire.Bind(new(branch.BranchService), new(*branch.BaseBranchService)),
	// 提供创建实现类实例的方法（关键修复）
	branch.NewBaseBranchService,
	wire.Bind(new(branchd.Repository), new(*branchd.MySQLRepository)),
	branchd.NewMySQLRepository,
)
