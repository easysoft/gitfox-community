package i18n

import (
	"os"
	"testing"
)

func TestFileLoader_LoadFromDir_RealLocalesDirectory(t *testing.T) {
	// 测试真实的 locales 目录
	localesDir := "./locales"

	// 检查目录是否存在
	if _, err := os.Stat(localesDir); os.IsNotExist(err) {
		t.Skip("Skipping test: locales directory does not exist")
	}

	loader := NewFileLoader()
	result, err := loader.LoadFromDir(localesDir)
	if err != nil {
		t.Fatalf("LoadFromDir failed on real locales directory: %v", err)
	}

	// 验证加载了预期的语言环境
	expectedLocales := []string{"en_US", "zh_CN"}
	for _, locale := range expectedLocales {
		if _, exists := result[locale]; !exists {
			t.Errorf("Expected locale '%s' to be loaded from real locales directory", locale)
		}
	}

	// 验证英文翻译内容
	if enTranslations, exists := result["en_US"]; exists {
		// 检查是否包含来自 space.json 和 repo.json 的翻译
		expectedKeys := map[string]string{
			"space.profile.title":      "Repository Profile",
			"space.profile.saveButton": "Save Changes",
			"space.settings.title":     "Repository Settings",
			"repo.profile.title":       "Repository Profile",
			"repo.profile.saveButton":  "Save Changes",
			"repo.settings.title":      "Repository Settings",
		}

		for key, expectedValue := range expectedKeys {
			if actualValue, exists := enTranslations[key]; !exists {
				t.Errorf("Expected key '%s' to exist in en_US translations from real locales", key)
			} else if actualValue != expectedValue {
				t.Errorf("Expected en_US['%s'] = '%s', got '%s'", key, expectedValue, actualValue)
			}
		}

		// 验证文件合并 - 应该包含来自多个文件的翻译
		if len(enTranslations) < 6 {
			t.Errorf("Expected at least 6 translation keys from merged files, got %d", len(enTranslations))
		}
	}

	// 验证中文翻译内容
	if zhTranslations, exists := result["zh_CN"]; exists {
		expectedKeys := map[string]string{
			"space.profile.title":      "项目资料",
			"space.profile.saveButton": "保存修改",
			"space.settings.title":     "项目设置",
			"repo.profile.title":       "仓库资料",
			"repo.profile.saveButton":  "保存修改",
			"repo.settings.title":      "仓库设置",
		}

		for key, expectedValue := range expectedKeys {
			if actualValue, exists := zhTranslations[key]; !exists {
				t.Errorf("Expected key '%s' to exist in zh_CN translations from real locales", key)
			} else if actualValue != expectedValue {
				t.Errorf("Expected zh_CN['%s'] = '%s', got '%s'", key, expectedValue, actualValue)
			}
		}
	}

	// 打印加载的翻译用于调试
	t.Logf("Loaded %d locales from real directory", len(result))
	for locale, translations := range result {
		t.Logf("Locale %s has %d translations", locale, len(translations))
		for key, value := range translations {
			t.Logf("  %s: %s", key, value)
		}
	}
}
