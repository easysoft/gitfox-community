package store

import (
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
)

type (
	// PrincipalInfoCache caches principal IDs to principalInfo.
	// PrincipalInfoCache 缓存 principalInfo
	PrincipalInfoCache cache.ExtendedCache[int64, *model.PrincipalInfo]
)
