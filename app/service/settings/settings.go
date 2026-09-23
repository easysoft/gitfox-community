// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package settings

import (
	"context"

	"github.com/easysoft/gitfox-community/app/constant"
	settingsModel "github.com/easysoft/gitfox-community/app/model/settings"
	"github.com/easysoft/gitfox-community/app/store/settings"
	"gorm.io/gorm"
)

type Service struct {
	settingsStore settings.Interface
}

// NewService 创建 SettingsService 实例
func NewSettingsService(settingsStore settings.Interface) Interface {
	return &Service{
		settingsStore: settingsStore,
	}
}

func (s *Service) NeedFitClient(ctx context.Context, repoID uint, key constant.SettingsKey) (bool, error) {
	setting, err := s.settingsStore.GetRepoSettingByKey(ctx, repoID, string(key))
	if err != nil && err != gorm.ErrRecordNotFound { // db error
		return false, err
	}

	if setting == nil { // not found return true
		return false, nil
	}

	if setting.Value != string(constant.DefaultRequireFitClient) {
		return true, nil
	}

	return false, nil
}

func (s *Service) GetSpaceSettings(ctx context.Context, spaceID uint) ([]*settingsModel.Setting, error) {
	return s.settingsStore.GetSpaceSettings(ctx, spaceID)
}

func (s *Service) GetRepoSettings(ctx context.Context, repoID uint) ([]*settingsModel.Setting, error) {
	return s.settingsStore.GetRepoSettings(ctx, repoID)
}

func (s *Service) GetSystemSettings(ctx context.Context) ([]*settingsModel.SystemSetting, error) {
	return s.settingsStore.GetSystemSettings(ctx)
}

func (s *Service) GetSpaceSettingByKey(ctx context.Context, spaceID uint, key string) (*settingsModel.Setting, error) {
	return s.settingsStore.GetSpaceSettingByKey(ctx, spaceID, key)
}

func (s *Service) GetRepoSettingByKey(ctx context.Context, repoID uint, key string) (*settingsModel.Setting, error) {
	return s.settingsStore.GetRepoSettingByKey(ctx, repoID, key)
}

func (s *Service) GetSystemSettingByKey(ctx context.Context, key string) (*settingsModel.SystemSetting, error) {
	return s.settingsStore.GetSystemSettingByKey(ctx, key)
}
