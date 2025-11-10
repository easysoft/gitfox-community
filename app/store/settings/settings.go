package settings

import (
	"context"

	SettingsModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/settings"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewSettingStore(db *gorm.DB) Interface {
	return &Store{db: db}
}

func (s *Store) GetSpaceSettings(ctx context.Context, spaceID int64) ([]*SettingsModel.Setting, error) {
	var out []*Settings
	if err := s.db.WithContext(ctx).
		Where("setting_space_id = ? AND deleted = 0", spaceID).
		Find(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return toSettings(out), nil
}

func (s *Store) GetRepoSettings(ctx context.Context, repoID int64) ([]*SettingsModel.Setting, error) {
	var out []*Settings
	if err := s.db.WithContext(ctx).
		Where("setting_repo_id = ? AND deleted = 0", repoID).
		Find(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return toSettings(out), nil
}

func (s *Store) GetSystemSettings(ctx context.Context) ([]*SettingsModel.SystemSetting, error) {
	var out []*Settings
	if err := s.db.WithContext(ctx).
		Where("setting_space_id = 0 AND setting_repo_id = 0 AND deleted = 0").
		Find(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return toSystemSettings(out), nil
}

func (s *Store) GetSpaceSettingByKey(ctx context.Context, spaceID int64, key string) (*SettingsModel.Setting, error) {
	var out Settings
	if err := s.db.WithContext(ctx).
		Where("setting_space_id = ? AND setting_key = ? AND deleted = 0", spaceID, key).
		First(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return toSetting(&out), nil
}

func (s *Store) GetRepoSettingByKey(ctx context.Context, repoID int64, key string) (*SettingsModel.Setting, error) {
	var out Settings
	if err := s.db.WithContext(ctx).
		Where("setting_repo_id = ? AND setting_key = ? AND deleted = 0", repoID, key).
		First(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return toSetting(&out), nil
}

func (s *Store) GetSystemSettingByKey(ctx context.Context, key string) (*SettingsModel.SystemSetting, error) {
	var out Settings
	if err := s.db.WithContext(ctx).
		Where("setting_space_id = 0 AND setting_repo_id = 0 AND setting_key = ? AND deleted = 0", key).
		First(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return toSystemSetting(&out), nil
}
