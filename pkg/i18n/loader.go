// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

// Loader 翻译文件加载器接口
type Loader interface {
	// LoadFromFile 从文件加载翻译
	LoadFromFile(filePath string) (locale string, translations map[string]string, err error)
	// LoadFromDir 从目录加载所有翻译文件
	LoadFromDir(dirPath string) (map[string]map[string]string, error)
	// LoadFromBytes 从字节数据加载翻译
	LoadFromBytes(data []byte, format string) (map[string]string, error)
}

// FileLoader 文件加载器
type FileLoader struct {
	supportedFormats map[string]FormatParser
}

// FormatParser 格式解析器
type FormatParser func(data []byte) (map[string]string, error)

// NewFileLoader 创建文件加载器
func NewFileLoader() Loader {
	return &FileLoader{
		supportedFormats: map[string]FormatParser{
			".json": parseJSON,
			".yaml": parseYAML,
		},
	}
}

// LoadFromFile 从文件加载翻译
func (l *FileLoader) LoadFromFile(filePath string) (locale string, translations map[string]string, err error) {
	// 从文件名提取语言环境
	locale = extractLocaleFromFilename(filePath)
	if locale == "" {
		return "", nil, errors.Errorf("cannot extract locale from filename: %s", filePath)
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to read file: %s", filePath)
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	parser, exists := l.supportedFormats[ext]
	if !exists {
		return "", nil, errors.Errorf("unsupported file format: %s", ext)
	}

	// 解析文件内容
	translations, err = parser(data)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to parse file: %s", filePath)
	}

	return locale, translations, nil
}

// LoadFromDir 从目录加载所有翻译文件
func (l *FileLoader) LoadFromDir(dirPath string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)

	// 遍历目录中的所有文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查是否为支持的文件格式
		ext := strings.ToLower(filepath.Ext(path))
		if _, supported := l.supportedFormats[ext]; !supported {
			return nil
		}

		// 加载翻译文件
		locale, translations, err := l.LoadFromFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to load translation file: %s", path)
		}

		// 合并翻译
		if result[locale] == nil {
			result[locale] = make(map[string]string)
		}
		for key, value := range translations {
			result[locale][key] = value
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed to walk directory: %s", dirPath)
	}

	return result, nil
}

// LoadFromBytes 从字节数据加载翻译
func (l *FileLoader) LoadFromBytes(data []byte, format string) (map[string]string, error) {
	parser, exists := l.supportedFormats[format]
	if !exists {
		return nil, errors.Errorf("unsupported format: %s", format)
	}

	return parser(data)
}

// extractLocaleFromFilename 从文件名或路径提取语言环境
func extractLocaleFromFilename(filename string) string {
	dir := filepath.Dir(filename)
	if dir != "." && dir != "/" {
		dirName := filepath.Base(dir)
		normalized := NormalizeLocale(dirName)
		if normalized != strings.ToLower(dirName) || isValidLocale(normalized) {
			return normalized
		}
	}

	// 然后尝试从文件名本身提取
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// 支持的文件名格式：
	// - en.json
	// - zh-CN.json
	// - zh_cn.json
	// - chinese.json
	// - messages.en.json
	// - app.zh-CN.json

	parts := strings.Split(nameWithoutExt, ".")
	if len(parts) >= 2 {
		// 最后一部分可能是语言环境
		lastPart := parts[len(parts)-1]
		normalized := NormalizeLocale(lastPart)
		if normalized != strings.ToLower(lastPart) || isValidLocale(normalized) {
			return normalized
		}
	}

	// 如果只有一个部分，检查是否为有效的语言环境
	if len(parts) == 1 {
		normalized := NormalizeLocale(parts[0])
		if normalized != strings.ToLower(parts[0]) || isValidLocale(normalized) {
			return normalized
		}
	}

	return ""
}

// isValidLocale 检查是否为有效的语言环境
func isValidLocale(locale string) bool {
	// 简单的语言环境验证
	// 支持格式：en, zh, zh-CN, zh-TW, en-US, zh_CN, en_US 等
	if len(locale) < 2 {
		return false
	}

	// 支持连字符和下划线分隔符
	parts := strings.FieldsFunc(locale, func(c rune) bool {
		return c == '-' || c == '_'
	})

	if len(parts) > 2 {
		return false
	}

	// 检查语言代码（2个字符）
	if len(parts[0]) != 2 {
		return false
	}

	// 如果有国家/地区代码，检查长度（2个字符）
	if len(parts) == 2 && len(parts[1]) != 2 {
		return false
	}

	return true
}

// parseJSON 解析 JSON 格式（支持嵌套结构）
func parseJSON(data []byte) (map[string]string, error) {
	var nested map[string]interface{}
	if err := json.Unmarshal(data, &nested); err != nil {
		return nil, errors.Wrap(err, "failed to parse JSON")
	}

	// 扁平化嵌套结构
	result := make(map[string]string)
	flattenMap(nested, "", result)
	return result, nil
}

// parseYAML 解析 YAML 格式（支持嵌套结构）
func parseYAML(data []byte) (map[string]string, error) {
	var nested map[string]interface{}
	if err := yaml.Unmarshal(data, &nested); err != nil {
		return nil, errors.Wrap(err, "failed to parse YAML")
	}

	// 扁平化嵌套结构
	result := make(map[string]string)
	flattenMap(nested, "", result)
	return result, nil
}

// flattenMap 扁平化嵌套的 map
func flattenMap(data map[string]interface{}, prefix string, result map[string]string) {
	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flattenMap(v, newKey, result)
		case string:
			result[newKey] = v
		}
	}
}
