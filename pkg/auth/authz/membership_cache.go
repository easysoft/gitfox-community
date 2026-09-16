package authz

import (
	"context"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"

	"github.com/easysoft/gitfox-community/pkg/cache"
)

type PermissionCacheKey struct {
	Account    string
	Scope      constant.Scope
	Permission constant.Permission
}

type PublicAccessCacheKey struct {
	Scope constant.Scope
}

type PermissionCache cache.Cache[PermissionCacheKey, bool]
type PublicAccessCache cache.Cache[PublicAccessCacheKey, bool]

func NewPermissionCache(cacheDuration time.Duration) PermissionCache {
	return cache.New(permissionCacheGetter{}, cacheDuration)
}

func NewPublicAccessCache(cacheDuration time.Duration) PublicAccessCache {
	return cache.New(publicAccessCacheGetter{}, cacheDuration)
}

type permissionCacheGetter struct {
}
type publicAccessCacheGetter struct {
}

// Find implements cache.Getter.
func (p permissionCacheGetter) Find(ctx context.Context, key PermissionCacheKey) (bool, error) {
	return true, nil
	// panic("unimplemented")
}

// Find implements cache.Getter.
func (p publicAccessCacheGetter) Find(ctx context.Context, key PublicAccessCacheKey) (bool, error) {
	return true, nil
	//panic("unimplemented")
}
