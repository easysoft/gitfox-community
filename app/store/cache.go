package store

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
)

type (
	// PrincipalInfoCache caches principal IDs to principal info.
	PrincipalInfoCache cache.ExtendedCache[int64, *model.PrincipalInfo]
)
