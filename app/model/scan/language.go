package scan

import "strings"

// NormalizeLang 将前端可能传入的 url 友好语言参数 (如 dotnet, csharp) 统一转换为底层存储使用的标准名称 (如 C#)
func NormalizeLang(lang string) string {
	lowerLang := strings.ToLower(lang)
	if lowerLang == "dotnet" || lowerLang == "csharp" {
		return "C#"
	}
	return lang
}

// NormalizeLangs 批量转换语言参数
func NormalizeLangs(langs []string) {
	for i, lang := range langs {
		langs[i] = NormalizeLang(lang)
	}
}
