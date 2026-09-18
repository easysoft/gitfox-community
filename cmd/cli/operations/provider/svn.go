package provider

import (
	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/app/api/controller/svn"
)

// SvnWireSet 仅装配 controller；底层 svnPkg.Service 已由 RepoWireSet → svnPkg.SvnWireSet 提供。
var SvnWireSet = wire.NewSet(
	svn.NewController,
)
