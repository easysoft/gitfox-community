package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/commit"
	commitSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/commit"
)

var CommitWireSet = wire.NewSet(
	commit.NewController,
	wire.Bind(new(commitSvc.Interface), new(*commitSvc.Service)),
	commitSvc.NewService,
)
