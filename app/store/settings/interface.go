package settings

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/settings"
)

type Interface interface {
	GetSpaceSettings(ctx context.Context, spaceID int) ([]*settings.Setting, error)
	GetRepoSettings(ctx context.Context, repoID int) ([]*settings.Setting, error)
	GetSystemSettings(ctx context.Context) ([]*settings.SystemSetting, error)

	GetSpaceSettingByKey(ctx context.Context, spaceID int, key string) (*settings.Setting, error)
	GetRepoSettingByKey(ctx context.Context, repoID int, key string) (*settings.Setting, error)
	GetSystemSettingByKey(ctx context.Context, key string) (*settings.SystemSetting, error)
}
