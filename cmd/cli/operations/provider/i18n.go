package provider

import (
	"os"
	"path/filepath"

	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils"
)

// I18nManager i18n管理器
type I18nManager struct {
	*i18n.Manager
}

var I18nWireSet = wire.NewSet(
	NewI18nManager,
)

// NewI18nManager 创建并初始化i18n管理器
func NewI18nManager(config *config.Config) (*I18nManager, error) {
	if config == nil {
		return nil, errors.SystemErr.New("i18n config is nil")
	}

	i18nConfig := config.I18n

	// 设置默认的 Header 名称，如果其他检测方式都未配置
	if i18nConfig.HeaderName == "" && i18nConfig.QueryParam == "" && i18nConfig.CookieName == "" && i18nConfig.SessionKey == "" {
		i18nConfig.HeaderName = i18n.XLocaleHeader
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, errors.SystemErr.Wrap(err, "[i18n]failed to get executable path")
	}
	exeDir := filepath.Dir(exePath)
	dstDir := filepath.Join(exeDir, "locales")

	if _, statErr := os.Stat(dstDir); os.IsNotExist(statErr) {
		if err := utils.CopyEmbeddedDir(resources.EmbeddedLocales, "locales", dstDir); err != nil {
			return nil, errors.SystemErr.Wrap(err, "[i18n]failed to copy embedded locales")
		}
	}

	i18n.InitGlobal(&i18n.ManagerConfig{
		DefaultLocale:    i18nConfig.DefaultLocale,
		FallbackLocale:   i18nConfig.FallbackLocale,
		TranslationsDir:  "",
		SupportedLocales: i18nConfig.SupportedLocales,
		AutoLoad:         true,
		HeaderName:       i18nConfig.HeaderName,
		QueryParam:       i18nConfig.QueryParam,
		CookieName:       i18nConfig.CookieName,
		SessionKey:       i18nConfig.SessionKey,
	})

	if err := os.RemoveAll(dstDir); err != nil {
		return nil, errors.SystemErr.Wrap(err, "[i18n]failed to remove temporary locales directory")
	}

	if i18nConfig.TranslationsDir != "" {
		i18n.GetGlobalManager().LoadTranslationsFromDir(i18nConfig.TranslationsDir)
	}

	return &I18nManager{i18n.GetGlobalManager()}, nil
}
