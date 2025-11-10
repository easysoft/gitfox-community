package settings

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/settings"
)

type Interface interface {
	NeedFitClient(ctx context.Context, repoID uint, key constant.SettingsKey) (bool, error)
	GetSpaceSettings(ctx context.Context, spaceID uint) ([]*settings.Setting, error)
	GetRepoSettings(ctx context.Context, repoID uint) ([]*settings.Setting, error)
	GetSystemSettings(ctx context.Context) ([]*settings.SystemSetting, error)

	GetSpaceSettingByKey(ctx context.Context, spaceID uint, key string) (*settings.Setting, error)
	GetRepoSettingByKey(ctx context.Context, repoID uint, key string) (*settings.Setting, error)
	GetSystemSettingByKey(ctx context.Context, key string) (*settings.SystemSetting, error)
}
