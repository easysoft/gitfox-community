package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/branch"
	branchSvc "github.com/easysoft/gitfox-community/app/service/branch"
	branchStore "github.com/easysoft/gitfox-community/app/store/branch"
)

var BranchWireSet = wire.NewSet(
	branch.NewController,
	wire.Bind(new(branchSvc.Interface), new(*branchSvc.Service)),
	branchSvc.NewService,
	wire.Bind(new(branchStore.Interface), new(*branchStore.Store)),
	branchStore.NewBranchStore,
)

// CommunityArchiveWireSet binds the community-edition no-op branch archive
// implementation. The paid edition replaces this with its own wire set.
var CommunityArchiveWireSet = wire.NewSet(
	wire.Bind(new(branchSvc.IArchiveService), new(*branchSvc.NoopArchiveService)),
	branchSvc.NewNoopArchiveService,
)
