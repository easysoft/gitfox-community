// 在 gitfox-dazu 项目中
package main

import (
	"fmt"
	"log"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
	// 获取所有翻译数据
	entries, err := i18n.GetProjectTranslations()
	if err != nil {
		log.Fatal("获取翻译失败:", err)
	}

	fmt.Printf("共获取到 %d 条翻译\n", len(entries))
	for _, entry := range entries {
		fmt.Printf("语言: %s, 键: %s, 值: %s\n",
			entry.Locale, entry.Key, entry.Value)
	}

	// 或者只获取中文翻译
	zhTranslations, err := i18n.GetProjectTranslationsByLocale("zh-CN")
	if err != nil {
		log.Fatal("获取中文翻译失败:", err)
	}

	fmt.Println("\n中文翻译:")
	for key, value := range zhTranslations {
		fmt.Printf("%s: %s\n", key, value)
	}

	// 获取支持的语言列表
	locales, err := i18n.GetProjectAvailableLocales()
	if err != nil {
		log.Fatal("获取语言列表失败:", err)
	}

	fmt.Printf("\n支持的语言: %v\n", locales)
}
