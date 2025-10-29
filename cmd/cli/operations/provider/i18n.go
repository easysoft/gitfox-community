package provider

import (
	"path/filepath"
	"runtime"

	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	pkg_errors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
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
		return nil, pkg_errors.SystemErr.New("i18n config is nil")
	}

	i18nConfig := config.I18n

	// 设置默认的 Header 名称，如果其他检测方式都未配置
	if i18nConfig.HeaderName == "" && i18nConfig.QueryParam == "" && i18nConfig.CookieName == "" && i18nConfig.SessionKey == "" {
		i18nConfig.HeaderName = i18n.XLocaleHeader
	}

	// todo 临时处理
	// 处理翻译文件目录路径，如果是相对路径，则相对于项目根目录
	translationsDir := i18nConfig.TranslationsDir
	if translationsDir != "" && !filepath.IsAbs(translationsDir) {
		// 获取项目根目录（相对于当前文件向上3级：cmd/cli/operations/provider -> 项目根目录）
		_, currentFile, _, _ := runtime.Caller(0)
		projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))))
		translationsDir = filepath.Join(projectRoot, translationsDir)
	}

	i18n.InitGlobal(&i18n.ManagerConfig{
		DefaultLocale:    i18nConfig.DefaultLocale,
		FallbackLocale:   i18nConfig.FallbackLocale,
		TranslationsDir:  translationsDir,
		SupportedLocales: i18nConfig.SupportedLocales,
		AutoLoad:         i18nConfig.AutoLoad,
		HeaderName:       i18nConfig.HeaderName,
		QueryParam:       i18nConfig.QueryParam,
		CookieName:       i18nConfig.CookieName,
		SessionKey:       i18nConfig.SessionKey,
	})

	return &I18nManager{i18n.GetGlobalManager()}, nil
}
