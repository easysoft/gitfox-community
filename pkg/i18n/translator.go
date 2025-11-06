package i18n

import (
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/pkg/errors"
)

// Translator 翻译器接口
type Translator interface {
	// T 翻译字符串
	T(key string, args ...interface{}) string
	// Tf 格式化翻译字符串
	Tf(key string, data interface{}) string
	// SetLocale 设置当前语言环境
	SetLocale(locale string) error
	// GetLocale 获取当前语言环境
	GetLocale() string
	// GetAvailableLocales 获取所有可用的语言环境
	GetAvailableLocales() []string
	// LoadTranslations 加载翻译文件
	LoadTranslations(locale string, data map[string]string) error
	// AddTranslation 添加单个翻译
	AddTranslation(locale, key, value string)
	// HasTranslation 检查是否存在翻译
	HasTranslation(locale, key string) bool
}

// translator 翻译器实现
type translator struct {
	mu              sync.RWMutex
	currentLocale   string
	defaultLocale   string
	fallbackLocale  string
	translations    map[string]map[string]string // locale -> key -> value
	missingHandler  MissingTranslationHandler
	templateFuncMap template.FuncMap
}

// PluralRule 复数规则函数
type PluralRule func(n int) int

// MissingTranslationHandler 缺失翻译处理器
type MissingTranslationHandler func(locale, key string) string

// TranslatorConfig 翻译器配置
type TranslatorConfig struct {
	CurrentLocale   string
	DefaultLocale   string
	FallbackLocale  string
	MissingHandler  MissingTranslationHandler
	TemplateFuncMap template.FuncMap
}

// NewTranslator 创建新的翻译器
func NewTranslator(config *TranslatorConfig) Translator {
	if config == nil {
		config = &TranslatorConfig{}
	}

	if config.CurrentLocale == "" {
		config.CurrentLocale = config.DefaultLocale
	}

	if config.DefaultLocale == "" {
		config.DefaultLocale = "zh-CN"
	}

	if config.FallbackLocale == "" {
		config.FallbackLocale = config.DefaultLocale
	}

	if config.MissingHandler == nil {
		config.MissingHandler = defaultMissingHandler
	}

	t := &translator{
		currentLocale:   config.CurrentLocale,
		defaultLocale:   config.DefaultLocale,
		fallbackLocale:  config.FallbackLocale,
		translations:    make(map[string]map[string]string),
		missingHandler:  config.MissingHandler,
		templateFuncMap: config.TemplateFuncMap,
	}

	return t
}

// T 翻译字符串
func (t *translator) T(key string, args ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 尝试当前语言环境
	if translation := t.getTranslation(t.currentLocale, key); translation != "" {
		return t.formatTranslation(translation, args...)
	}

	// 尝试回退语言环境
	if t.currentLocale != t.fallbackLocale {
		if translation := t.getTranslation(t.fallbackLocale, key); translation != "" {
			return t.formatTranslation(translation, args...)
		}
	}

	// 使用缺失处理器
	return t.missingHandler(t.currentLocale, key)
}

// Tf 格式化翻译字符串（使用模板）
func (t *translator) Tf(key string, data interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 尝试当前语言环境
	if translation := t.getTranslation(t.currentLocale, key); translation != "" {
		return t.executeTemplate(translation, data)
	}

	// 尝试回退语言环境
	if t.currentLocale != t.fallbackLocale {
		if translation := t.getTranslation(t.fallbackLocale, key); translation != "" {
			return t.executeTemplate(translation, data)
		}
	}

	return t.missingHandler(t.currentLocale, key)
}

// SetLocale 设置当前语言环境
func (t *translator) SetLocale(locale string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.translations[locale]; !exists && locale != t.defaultLocale {
		return errors.Errorf("locale %s not found", locale)
	}

	t.currentLocale = locale
	return nil
}

// GetLocale 获取当前语言环境
func (t *translator) GetLocale() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLocale
}

// GetAvailableLocales 获取所有可用的语言环境
func (t *translator) GetAvailableLocales() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	locales := make([]string, 0, len(t.translations))
	for locale := range t.translations {
		locales = append(locales, locale)
	}
	return locales
}

// LoadTranslations 加载翻译文件
func (t *translator) LoadTranslations(locale string, data map[string]string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.translations[locale] == nil {
		t.translations[locale] = make(map[string]string)
	}

	for key, value := range data {
		t.translations[locale][key] = value
	}

	return nil
}

// AddTranslation 添加单个翻译
func (t *translator) AddTranslation(locale, key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.translations[locale] == nil {
		t.translations[locale] = make(map[string]string)
	}

	t.translations[locale][key] = value
}

// HasTranslation 检查是否存在翻译
func (t *translator) HasTranslation(locale, key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if translations, exists := t.translations[locale]; exists {
		_, exists := translations[key]
		return exists
	}
	return false
}

// getTranslation 获取翻译（内部方法，不加锁）
func (t *translator) getTranslation(locale, key string) string {
	if translations, exists := t.translations[locale]; exists {
		if translation, exists := translations[key]; exists {
			return translation
		}
	}
	return ""
}

// formatTranslation 格式化翻译字符串
func (t *translator) formatTranslation(translation string, args ...interface{}) string {
	if len(args) == 0 {
		return translation
	}

	// 如果翻译中没有格式化占位符，直接返回，避免 %!(EXTRA ...)
	if !strings.Contains(translation, "%") {
		return translation
	}

	// 支持 fmt.Sprintf 风格的格式化
	return fmt.Sprintf(translation, args...)
}

// executeTemplate 执行模板
func (t *translator) executeTemplate(translation string, data interface{}) string {
	tmpl, err := template.New("translation").Funcs(t.templateFuncMap).Parse(translation)
	if err != nil {
		return translation // 如果模板解析失败，返回原始字符串
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return translation // 如果模板执行失败，返回原始字符串
	}

	return buf.String()
}

// defaultMissingHandler 默认的缺失翻译处理器
func defaultMissingHandler(locale, key string) string {
	return fmt.Sprintf("[%s:%s]", locale, key)
}
