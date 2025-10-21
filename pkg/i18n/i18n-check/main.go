package main

import (
	"fmt"
	"os"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run pkg/i18n/i18n-check/main.go <locales-dir>")
		fmt.Println("Example: go run pkg/i18n/i18n-check/main.go locales")
		os.Exit(1)
	}

	localesDir := os.Args[1]

	fmt.Printf("🔍 检查多语言文件键一致性: %s\n", localesDir)

	// 直接使用现有的 Manager 和 LoadTranslationsFromDir 方法
	// 该方法内部已经集成了 validateKeyConsistency 验证
	manager := i18n.NewManager(&i18n.ManagerConfig{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
	})

	err := manager.LoadTranslationsFromDir(localesDir)
	if err != nil {
		fmt.Printf("❌ 键一致性检查失败:\n%s\n", err.Error())
		os.Exit(1)
	}

	// 获取统计信息
	locales := manager.GetAvailableLocales()
	fmt.Printf("✅ 键一致性检查通过！\n")
	fmt.Printf("📊 检查的语言环境: %v\n", locales)

	// 统计每个语言环境的翻译数量
	for _, locale := range locales {
		_, err := manager.GetTranslator(locale)
		if err != nil {
			continue
		}

		// 通过重新加载来获取原始数据进行统计
		loader := i18n.NewFileLoader()
		translations, err := loader.LoadFromDir(localesDir)
		if err != nil {
			continue
		}

		if localeTranslations, exists := translations[locale]; exists {
			fmt.Printf("  %s: %d 个翻译键\n", locale, len(localeTranslations))
		}
	}

	fmt.Printf("🎉 所有翻译键在各语言环境中保持一致！\n")
}
