package i18n

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// ValidateTranslations 验证翻译完整性
func ValidateTranslations(baseFile string, targetFiles []string) (*ValidationResult, error) {
	// 加载基础翻译文件
	baseTranslations, err := loadJSONFile(baseFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load base translation file: %s", baseFile)
	}

	result := &ValidationResult{
		BaseFile:    baseFile,
		BaseKeys:    len(baseTranslations),
		FileResults: make(map[string]*FileValidationResult),
	}

	// 验证每个目标文件
	for _, targetFile := range targetFiles {
		targetTranslations, err := loadJSONFile(targetFile)
		if err != nil {
			result.FileResults[targetFile] = &FileValidationResult{
				Error: err.Error(),
			}
			continue
		}

		fileResult := &FileValidationResult{
			TotalKeys:   len(targetTranslations),
			MissingKeys: []string{},
			ExtraKeys:   []string{},
		}

		// 检查缺失的键
		for key := range baseTranslations {
			if _, exists := targetTranslations[key]; !exists {
				fileResult.MissingKeys = append(fileResult.MissingKeys, key)
			}
		}

		// 检查多余的键
		for key := range targetTranslations {
			if _, exists := baseTranslations[key]; !exists {
				fileResult.ExtraKeys = append(fileResult.ExtraKeys, key)
			}
		}

		fileResult.Coverage = float64(fileResult.TotalKeys-len(fileResult.MissingKeys)) / float64(result.BaseKeys) * 100

		result.FileResults[targetFile] = fileResult
	}

	return result, nil
}

// ValidationResult 验证结果
type ValidationResult struct {
	BaseFile    string                           `json:"base_file"`
	BaseKeys    int                              `json:"base_keys"`
	FileResults map[string]*FileValidationResult `json:"file_results"`
}

// FileValidationResult 文件验证结果
type FileValidationResult struct {
	TotalKeys   int      `json:"total_keys"`
	MissingKeys []string `json:"missing_keys"`
	ExtraKeys   []string `json:"extra_keys"`
	Coverage    float64  `json:"coverage"`
	Error       string   `json:"error,omitempty"`
}

// loadJSONFile 加载 JSON 文件
func loadJSONFile(filename string) (map[string]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return nil, err
	}

	return translations, nil
}

// NormalizeLocale 规范化语言环境
func NormalizeLocale(locale string) string {
	// 转换为小写
	locale = strings.ToLower(locale)

	// 处理常见的变体
	switch locale {
	case "zh_cn", "zh-cn", "chinese", "cn", "zh":
		return "zh-CN"
	case "zh_tw", "zh-tw", "traditional chinese", "tw":
		return "zh-TW"
	case "english", "eng":
		return "en"
	case "en_us", "en-us":
		return "en-US"
	case "en_gb", "en-gb":
		return "en-GB"
	}

	return locale
}

// GetSupportedLocales 获取支持的语言环境列表
func GetSupportedLocales() []string {
	return []string{
		"en",    // English
		"zh-CN", // Simplified Chinese
		"zh-TW", // Traditional Chinese
		"ja",    // Japanese
		"ko",    // Korean
		"fr",    // French
		"de",    // German
		"es",    // Spanish
		"pt",    // Portuguese
		"ru",    // Russian
	}
}
