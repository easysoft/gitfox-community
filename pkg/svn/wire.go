package svn

import (
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/google/wire"
)

// ProvideSVN returns a new SVN Service implementing Interface.
func ProvideSVN(cfg *config.Config) (Interface, error) {
	return New(cfg)
}

// SvnWireSet provides the SVN service binding.
var SvnWireSet = wire.NewSet(
	ProvideSVN,
)
