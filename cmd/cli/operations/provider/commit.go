package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/commit"
	commitSvc "github.com/easysoft/gitfox-community/app/service/commit"
)

var CommitWireSet = wire.NewSet(
	commit.NewController,
	wire.Bind(new(commitSvc.Interface), new(*commitSvc.Service)),
	commitSvc.NewService,
)
