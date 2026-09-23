// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package i18n

import (
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

// TranslationEntry 翻译条目
type TranslationEntry struct {
	Locale string `json:"locale"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

// GetProjectTranslations 获取 gitfox-community 项目内置的翻译数据
// 外部项目可以通过此方法获取项目的中英文翻译
func GetProjectTranslations() ([]TranslationEntry, error) {
	// 获取当前文件的路径
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("failed to get current file path")
	}

	// 计算 locales 目录的路径
	// 从 pkg/i18n/manager.go 向上两级到项目根目录，然后进入 locales
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	localesDir := filepath.Join(projectRoot, "locales")

	// 使用 loader 直接加载翻译文件
	loader := NewFileLoader()
	translations, err := loader.LoadFromDir(localesDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load gitfox-community translations")
	}

	// 转换为 TranslationEntry 切片
	var entries []TranslationEntry
	for locale, localeTranslations := range translations {
		for key, value := range localeTranslations {
			entries = append(entries, TranslationEntry{
				Locale: locale,
				Key:    key,
				Value:  value,
			})
		}
	}

	return entries, nil
}

// GetProjectTranslationsByLocale 按语言环境获取翻译数据
func GetProjectTranslationsByLocale(locale string) (map[string]string, error) {
	entries, err := GetProjectTranslations()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, entry := range entries {
		if entry.Locale == locale {
			result[entry.Key] = entry.Value
		}
	}

	return result, nil
}

// GetProjectAvailableLocales 获取项目支持的语言环境列表
func GetProjectAvailableLocales() ([]string, error) {
	entries, err := GetProjectTranslations()
	if err != nil {
		return nil, err
	}

	localeSet := make(map[string]bool)
	for _, entry := range entries {
		localeSet[entry.Locale] = true
	}

	var locales []string
	for locale := range localeSet {
		locales = append(locales, locale)
	}

	return locales, nil
}
