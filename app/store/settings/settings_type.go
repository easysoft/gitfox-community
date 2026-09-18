package settings

import (
	"time"

	SettingsModel "github.com/easysoft/gitfox-community/app/model/settings"
)

type Settings struct {
	ID          int        `gorm:"primaryKey"`
	SpaceID     uint       `gorm:"column:setting_space_id"`
	RepoID      uint       `gorm:"column:setting_repo_id"`
	Key         string     `gorm:"column:setting_key"`
	Value       string     `gorm:"column:setting_value"`
	CreatedBy   string     `gorm:"column:createdBy"`
	CreatedDate *time.Time `gorm:"column:createdDate"` // 允许为NULL，指针类型
	EditedBy    string     `gorm:"column:editedBy"`
	EditedDate  *time.Time `gorm:"column:editedDate"` // 允许为NULL，指针类型
	Deleted     uint8      `gorm:"column:deleted"`
}

func (Settings) TableName() string {
	return "settings"
}

func toSystemSetting(setting *Settings) *SettingsModel.SystemSetting {
	return &SettingsModel.SystemSetting{
		Key:   setting.Key,
		Value: setting.Value,
	}
}

func toSystemSettings(setting []*Settings) []*SettingsModel.SystemSetting {
	var out []*SettingsModel.SystemSetting
	for _, setting := range setting {
		out = append(out, toSystemSetting(setting))
	}
	return out
}

func toSetting(setting *Settings) *SettingsModel.Setting {
	return &SettingsModel.Setting{
		Key:   setting.Key,
		Value: setting.Value,
	}
}

func toSettings(settings []*Settings) []*SettingsModel.Setting {
	var out []*SettingsModel.Setting
	for _, setting := range settings {
		out = append(out, toSetting(setting))
	}
	return out
}
