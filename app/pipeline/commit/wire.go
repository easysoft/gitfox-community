package commit

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideService,
)

// ProvideService provides a service which can fetch commit
// information about a repository.
func ProvideService(git git.Interface) Service {
	return newService(git)
}
