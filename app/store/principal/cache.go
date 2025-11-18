package principal

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
)

type (
	// PrincipalInfoCache caches principal IDs to principalInfo.
	// PrincipalInfoCache 缓存 principalInfo
	PrincipalInfoCache cache.ExtendedCache[int64, *principal.PrincipalInfo]
)

// NewPrincipalInfoCache creates a cache for storing principal.PrincipalInfo objects.
func NewPrincipalInfoCache() PrincipalInfoCache {
	getter := PrincipalInfoCacheGetter{}
	return cache.NewExtended(getter, 30*time.Second)
}

type PrincipalInfoCacheGetter struct {
	// devops client
}

func (g PrincipalInfoCacheGetter) Find(ctx context.Context, key int64) (*principal.PrincipalInfo, error) {
	return nil, nil
}

func (g PrincipalInfoCacheGetter) FindMany(ctx context.Context, keys []int64) ([]*principal.PrincipalInfo, error) {
	return nil, nil
}
