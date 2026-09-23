// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/i18n"
)

type Config struct {
	LocalesDir  string
	OutputFile  string
	PackageName string
	CheckOnly   bool
	ConstPrefix string
}

type KeyInfo struct {
	Key         string
	ConstName   string
	FullKey     string
	Description string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	config := parseArgs()

	fmt.Printf("🔍 扫描 locales 目录: %s\n", config.LocalesDir)

	// 1. 收集所有的 keys
	allKeys, err := collectAllKeys(config.LocalesDir)
	if err != nil {
		fmt.Printf("❌ 收集 keys 失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 检查 key 一致性
	if err := checkKeyConsistency(config.LocalesDir, allKeys); err != nil {
		fmt.Printf("❌ Key 一致性检查失败:\n%s\n", err)
		if config.CheckOnly {
			os.Exit(1)
		}
		fmt.Println("⚠️  继续生成常量，但建议先修复一致性问题")
	} else {
		fmt.Println("✅ Key 一致性检查通过")
	}

	if config.CheckOnly {
		fmt.Printf("📊 总共找到 %d 个 keys\n", len(allKeys))
		return
	}

	// 3. 生成常量
	if err := generateConstants(config, allKeys); err != nil {
		fmt.Printf("❌ 生成常量失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 常量生成完成: %s\n", config.OutputFile)
	fmt.Printf("📊 总共生成 %d 个常量\n", len(allKeys))
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/i18n-keygen/main.go [options] <locales-dir>")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -o, --output <file>     输出文件路径 (默认: pkg/i18n/i18n_keys.go)")
	fmt.Println("  -p, --package <name>    包名 (默认: i18n)")
	fmt.Println("  -c, --check             仅检查一致性，不生成文件")
	fmt.Println("  --prefix <prefix>       常量前缀 (默认: Key)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/i18n-keygen/main.go locales")
	fmt.Println("  go run cmd/i18n-keygen/main.go -c locales")
	fmt.Println("  go run cmd/i18n-keygen/main.go -o keys.go locales")
}

func parseArgs() *Config {
	config := &Config{
		OutputFile:  "pkg/i18n/i18n_keys.go",
		PackageName: "i18n",
		ConstPrefix: "Key",
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 >= len(args) {
				fmt.Println("❌ -o/--output 需要指定文件路径")
				os.Exit(1)
			}
			config.OutputFile = args[i+1]
			i++
		case "-p", "--package":
			if i+1 >= len(args) {
				fmt.Println("❌ -p/--package 需要指定包名")
				os.Exit(1)
			}
			config.PackageName = args[i+1]
			i++
		case "-c", "--check":
			config.CheckOnly = true
		case "--prefix":
			if i+1 >= len(args) {
				fmt.Println("❌ --prefix 需要指定前缀")
				os.Exit(1)
			}
			config.ConstPrefix = args[i+1]
			i++
		default:
			if !strings.HasPrefix(args[i], "-") {
				config.LocalesDir = args[i]
			}
		}
	}

	if config.LocalesDir == "" {
		fmt.Println("❌ 请指定 locales 目录")
		printUsage()
		os.Exit(1)
	}

	return config
}

func collectAllKeys(localesDir string) ([]KeyInfo, error) {
	// 使用 i18n 包的 FileLoader 来加载所有翻译文件
	loader := i18n.NewFileLoader()
	allTranslations, err := loader.LoadFromDir(localesDir)
	if err != nil {
		return nil, fmt.Errorf("加载翻译文件失败: %v", err)
	}

	// 收集所有唯一的 keys
	keySet := make(map[string]bool)
	for _, translations := range allTranslations {
		for key := range translations {
			keySet[key] = true
		}
	}

	// 转换为 KeyInfo 切片
	var allKeys []KeyInfo
	for fullKey := range keySet {
		// 优先使用中文翻译作为描述，如果没有则使用其他语言
		var description string

		// 首先尝试获取中文翻译
		if zhTranslations, exists := allTranslations["zh-CN"]; exists {
			if desc, exists := zhTranslations[fullKey]; exists {
				description = desc
			}
		}

		// // 如果没有中文翻译，则使用第一个可用的翻译
		if description == "" {
			for _, translations := range allTranslations {
				if desc, exists := translations[fullKey]; exists {
					description = desc
					break
				}
			}
		}

		parts := strings.Split(fullKey, ".")
		key := parts[len(parts)-1]

		allKeys = append(allKeys, KeyInfo{
			Key:         key,
			ConstName:   generateConstName(fullKey),
			FullKey:     fullKey,
			Description: description,
		})
	}

	// 排序
	sort.Slice(allKeys, func(i, j int) bool {
		return allKeys[i].FullKey < allKeys[j].FullKey
	})

	return allKeys, nil
}

func generateConstName(fullKey string) string {
	parts := strings.Split(fullKey, ".")
	var constParts []string

	for _, part := range parts {
		// 转换为驼峰命名
		constPart := strings.Title(strings.ReplaceAll(part, "_", " "))
		constPart = strings.ReplaceAll(constPart, " ", "")
		constParts = append(constParts, constPart)
	}

	return strings.Join(constParts, "")
}

func checkKeyConsistency(localesDir string, allKeys []KeyInfo) error {
	// 使用 i18n 包的 FileLoader 来加载所有翻译文件
	loader := i18n.NewFileLoader()
	allTranslations, err := loader.LoadFromDir(localesDir)
	if err != nil {
		return fmt.Errorf("加载翻译文件失败: %v", err)
	}

	// 构建所有 keys 的集合
	allKeySet := make(map[string]bool)
	for _, key := range allKeys {
		allKeySet[key.FullKey] = true
	}

	// 检查每个语言环境的一致性
	var inconsistencies []string
	for locale, translations := range allTranslations {
		// 检查缺失的 keys
		for fullKey := range allKeySet {
			if _, exists := translations[fullKey]; !exists {
				inconsistencies = append(inconsistencies, fmt.Sprintf("  - %s: 缺少 key '%s'", locale, fullKey))
			}
		}

		// 检查多余的 keys
		for fullKey := range translations {
			if !allKeySet[fullKey] {
				inconsistencies = append(inconsistencies, fmt.Sprintf("  - %s: 多余 key '%s'", locale, fullKey))
			}
		}
	}

	if len(inconsistencies) > 0 {
		return fmt.Errorf("发现 key 不一致:\n%s", strings.Join(inconsistencies, "\n"))
	}

	return nil
}

func generateConstants(config *Config, keys []KeyInfo) error {
	var sb strings.Builder

	// 包声明
	sb.WriteString(fmt.Sprintf("package %s\n\n", config.PackageName))

	// 生成注释
	sb.WriteString("// 此文件由 i18n-keygen 工具自动生成，请勿手动修改\n")
	sb.WriteString("// Generated by i18n-keygen tool, do not edit manually\n\n")

	// 定义 I18nKey 类型
	sb.WriteString("type I18nKey string\n\n")

	// 常量声明开始
	sb.WriteString("const (\n")

	// 按分组生成常量
	currentGroup := ""
	for _, key := range keys {
		group := strings.Split(key.FullKey, ".")[0]

		if group != currentGroup {
			if currentGroup != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("\t// %s 相关常量\n", strings.Title(group)))
			currentGroup = group
		}

		// 清理描述中的中文标点符号，保留中文字符
		description := key.Description

		// 先替换中文标点符号为英文标点符号
		var cleanDesc strings.Builder
		for _, r := range description {
			switch r {
			case 0x3002: // 中文句号 。
				cleanDesc.WriteRune('.')
			case 0xff0c: // 中文逗号 ，
				cleanDesc.WriteRune(',')
			case 0xff1a: // 中文冒号 ：
				cleanDesc.WriteRune(':')
			case 0xff01: // 中文感叹号 ！
				cleanDesc.WriteRune('!')
			case 0xff1f: // 中文问号 ？
				cleanDesc.WriteRune('?')
			case 0xff08: // 中文左括号 （
				cleanDesc.WriteRune('(')
			case 0xff09: // 中文右括号 ）
				cleanDesc.WriteRune(')')
			case 0x201c, 0x201d: // 中文引号 " "
				cleanDesc.WriteRune('"')
			default:
				// 保留其他字符（包括中文字符）
				cleanDesc.WriteRune(r)
			}
		}
		cleanedDescription := strings.TrimSpace(cleanDesc.String())

		// 按 Unicode 字符数量截断，避免截断中文字符
		runes := []rune(cleanedDescription)
		if len(runes) > 50 {
			cleanedDescription = string(runes[:47]) + "..."
		}

		// 使用自定义类型 I18nKey 生成常量
		sb.WriteString(fmt.Sprintf("\t%s I18nKey = \"%s\" // %s\n",
			key.ConstName, key.FullKey, cleanedDescription))
	}

	sb.WriteString(")\n")

	// 格式化代码
	formatted, err := format.Source([]byte(sb.String()))
	if err != nil {
		return fmt.Errorf("格式化代码失败: %v", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(config.OutputFile), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 写入文件
	return os.WriteFile(config.OutputFile, formatted, 0644)
}
