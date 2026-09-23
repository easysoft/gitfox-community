// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package user

import (
	"context"
	"time"

	principal "github.com/easysoft/gitfox-community/app/model/user"
	"github.com/easysoft/gitfox-community/pkg/cache"
)

type (
	// UserInfoCache caches user account to user.
	// UserInfoCache 缓存 user.Account 到 user 的映射
	UserInfoCache cache.ExtendedCache[string, *principal.User]
)

// NewUserInfoCache creates a cache for storing user.User objects.
func NewUserInfoCache() UserInfoCache {
	getter := UserInfoCacheGetter{}
	return cache.NewExtended(getter, 30*time.Second)
}

type UserInfoCacheGetter struct {
	// devops client
}

func (g UserInfoCacheGetter) Find(ctx context.Context, account string) (*principal.User, error) {
	return nil, nil
}

func (g UserInfoCacheGetter) FindMany(ctx context.Context, accounts []string) ([]*principal.User, error) {
	return nil, nil
}
