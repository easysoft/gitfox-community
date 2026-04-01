package cli

import (
	"fmt"
	"os"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
)

var currentLocale string

// Init 初始化 CLI i18n，根据系统环境自动检测语言并从内存加载翻译
//
// 性能优化：直接从 embedded resources 读取到内存，无需创建临时文件
func Init() {
	currentLocale = detectLocale()

	// 直接从 embedded FS 加载翻译到内存，无需写入文件系统
	translations, err := loadTranslationsFromEmbedFS()
	if err != nil {
		// 加载失败，但不中断程序
		return
	}

	// 初始化全局管理器（不设置 TranslationsDir，手动加载翻译）
	i18n.InitGlobal(&i18n.ManagerConfig{
		DefaultLocale:    "en-US",
		FallbackLocale:   "en-US",
		TranslationsDir:  "", // 不从目录加载
		SupportedLocales: []string{"en-US", "zh-CN"},
		AutoLoad:         false, // 手动加载
	})

	// 手动加载翻译到 Manager
	manager := i18n.GetGlobalManager()
	for locale, trans := range translations {
		for key, value := range trans {
			manager.AddTranslation(locale, key, value)
		}
	}

	// 返回空清理函数（无需清理，因为没有创建临时文件）
	return
}

// loadTranslationsFromEmbedFS 从 embedded FS 加载 CLI 翻译（仅加载 cli.yaml）
func loadTranslationsFromEmbedFS() (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	loader := i18n.NewFileLoader()

	// 支持的语言列表
	locales := []string{"en-US", "zh-CN"}

	for _, locale := range locales {
		// 构造文件路径：locales/{locale}/cli.yaml
		filePath := fmt.Sprintf("locales/%s/cli.yaml", locale)

		// 读取文件内容
		data, err := resources.EmbeddedLocales.ReadFile(filePath)
		if err != nil {
			// 如果文件不存在，跳过该语言
			continue
		}

		// 解析翻译（.yaml 扩展名）
		translations, err := loader.LoadFromBytes(data, ".yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}

		// 保存翻译
		result[locale] = translations
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no CLI translations found")
	}

	return result, nil
}

// detectLocale 检测系统语言环境
func detectLocale() string {
	// 优先级：GITFOX_LANG > LANG > LC_ALL
	if lang := os.Getenv("GITFOX_LANG"); lang != "" {
		return normalizeLocale(lang)
	}

	if lang := os.Getenv("LANG"); lang != "" {
		return normalizeLocale(lang)
	}

	if lang := os.Getenv("LC_ALL"); lang != "" {
		return normalizeLocale(lang)
	}

	// 默认英文
	return "en-US"
}

// normalizeLocale 标准化语言环境字符串
// 例如：zh_CN.UTF-8 -> zh-CN, en_US -> en-US
func normalizeLocale(locale string) string {
	// 移除编码部分 (如 .UTF-8)
	if idx := strings.Index(locale, "."); idx != -1 {
		locale = locale[:idx]
	}

	// 转换下划线为连字符
	locale = strings.Replace(locale, "_", "-", 1)

	// 检查是否为中文
	if strings.HasPrefix(strings.ToLower(locale), "zh") {
		return "zh-CN"
	}

	// 其他语言默认英文
	return "en-US"
}

// T 翻译函数，使用当前检测到的语言环境
func T(key i18n.I18nKey) string {
	return i18n.TWithLocale(currentLocale, string(key))
}
