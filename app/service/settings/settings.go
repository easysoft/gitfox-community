package settings

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	settingsModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/settings"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/settings"
	"gorm.io/gorm"
)

type Service struct {
	settingsStore settings.Interface
}

// NewService 创建 SpaceService 实例
func NewSpaceSvc(settingsStore settings.Interface) Interface {
	return &Service{
		settingsStore: settingsStore,
	}
}

func (s *Service) NeedFitClient(ctx context.Context, repoID int64, key constant.SettingsKey) (bool, error) {
	setting, err := s.settingsStore.GetRepoSettingByKey(ctx, repoID, string(key))
	if err != nil && err != gorm.ErrRecordNotFound { // db error
		return false, err
	}

	if setting == nil { // not found return true
		return true, nil
	}

	if setting.Value == string(constant.SettingsValueTrue) {
		return true, nil
	}

	return false, nil
}

func (s *Service) GetSpaceSettings(ctx context.Context, spaceID int64) ([]*settingsModel.Setting, error) {
	return s.settingsStore.GetSpaceSettings(ctx, spaceID)
}

func (s *Service) GetRepoSettings(ctx context.Context, repoID int64) ([]*settingsModel.Setting, error) {
	return s.settingsStore.GetRepoSettings(ctx, repoID)
}

func (s *Service) GetSystemSettings(ctx context.Context) ([]*settingsModel.SystemSetting, error) {
	return s.settingsStore.GetSystemSettings(ctx)
}

func (s *Service) GetSpaceSettingByKey(ctx context.Context, spaceID int64, key string) (*settingsModel.Setting, error) {
	return s.settingsStore.GetSpaceSettingByKey(ctx, spaceID, key)
}

func (s *Service) GetRepoSettingByKey(ctx context.Context, repoID int64, key string) (*settingsModel.Setting, error) {
	return s.settingsStore.GetRepoSettingByKey(ctx, repoID, key)
}

func (s *Service) GetSystemSettingByKey(ctx context.Context, key string) (*settingsModel.SystemSetting, error) {
	return s.settingsStore.GetSystemSettingByKey(ctx, key)
}
