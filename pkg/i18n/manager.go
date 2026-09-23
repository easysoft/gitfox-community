// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package i18n

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Manager i18n 管理器
type Manager struct {
	mu          sync.RWMutex
	translators map[string]Translator
	loader      Loader
	config      *ManagerConfig
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	DefaultLocale   string
	FallbackLocale  string
	TranslationsDir string
	AutoLoad        bool
	MissingHandler  MissingTranslationHandler

	SupportedLocales []string
	HeaderName       string
	QueryParam       string
	CookieName       string
	SessionKey       string
}

// NewManager 创建 i18n 管理器
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = &ManagerConfig{}
	}

	if config.DefaultLocale == "" {
		config.DefaultLocale = "zh-CN"
	}

	if config.FallbackLocale == "" {
		config.FallbackLocale = "en-US"
	}

	if config.TranslationsDir == "" {
		config.TranslationsDir = "locales"
	}

	m := &Manager{
		translators: make(map[string]Translator),
		loader:      NewFileLoader(),
		config:      config,
	}

	// 自动加载翻译文件
	if config.AutoLoad && config.TranslationsDir != "" {
		err := m.LoadTranslationsFromDir(config.TranslationsDir)
		if err != nil {
			panic("failed to load translations from directory: " + config.TranslationsDir)
		}
	}

	return m
}

func (m *Manager) ToMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		DefaultLocale:   m.config.DefaultLocale,
		AcceptedLocales: m.config.SupportedLocales,
		HeaderName:      m.config.HeaderName,
		QueryParam:      m.config.QueryParam,
		CookieName:      m.config.CookieName,
		SessionKey:      m.config.SessionKey,
	}
}

// GetTranslator 获取指定语言环境的翻译器
func (m *Manager) GetTranslator(locale string) (Translator, error) {
	m.mu.RLock()
	if translator, exists := m.translators[locale]; exists {
		m.mu.RUnlock()
		return translator, nil
	}
	m.mu.RUnlock()

	// 创建新的翻译器
	return m.createTranslator(locale)
}

// GetDefaultTranslator 获取默认翻译器
func (m *Manager) GetDefaultTranslator() Translator {
	translator, _ := m.GetTranslator(m.config.DefaultLocale)
	return translator
}

// LoadTranslationsFromDir 从目录加载翻译文件
func (m *Manager) LoadTranslationsFromDir(dirPath string) error {
	translations, err := m.loader.LoadFromDir(dirPath)
	if err != nil {
		return errors.Wrapf(err, "failed to load translations from directory: %s", dirPath)
	}

	// 校验键一致性
	if err := m.validateKeyConsistency(translations); err != nil {
		return errors.Wrap(err, "translation key consistency validation failed")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for locale, localeTranslations := range translations {
		translator := m.getOrCreateTranslator(locale)
		if err := translator.LoadTranslations(locale, localeTranslations); err != nil {
			return errors.Wrapf(err, "failed to load translations for locale: %s", locale)
		}
	}

	return nil
}

// validateKeyConsistency 校验所有语言环境的键一致性
func (m *Manager) validateKeyConsistency(translations map[string]map[string]string) error {
	if len(translations) <= 1 {
		return nil // 只有一个或没有语言环境，无需校验
	}

	// 收集所有键和对应的语言环境
	keyToLocales := make(map[string][]string)
	allLocales := make([]string, 0, len(translations))

	for locale, trans := range translations {
		allLocales = append(allLocales, locale)
		for key := range trans {
			keyToLocales[key] = append(keyToLocales[key], locale)
		}
	}

	var missingKeys []string
	totalLocales := len(allLocales)

	// 检查每个键是否在所有语言环境中都存在
	for key, locales := range keyToLocales {
		if len(locales) != totalLocales {
			// 找出缺失该键的语言环境
			localeSet := make(map[string]bool)
			for _, locale := range locales {
				localeSet[locale] = true
			}

			var missingLocales []string
			for _, locale := range allLocales {
				if !localeSet[locale] {
					missingLocales = append(missingLocales, locale)
				}
			}

			missingKeys = append(missingKeys,
				fmt.Sprintf("key '%s' missing in: %s", key, strings.Join(missingLocales, ", ")))
		}
	}

	if len(missingKeys) > 0 {
		return errors.New("translation key inconsistencies found:\n  - " +
			strings.Join(missingKeys, "\n  - "))
	}

	return nil
}

// AddTranslation 添加翻译
func (m *Manager) AddTranslation(locale, key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	translator := m.getOrCreateTranslator(locale)
	translator.AddTranslation(locale, key, value)
}

// GetAvailableLocales 获取所有可用的语言环境
func (m *Manager) GetAvailableLocales() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	locales := make([]string, 0, len(m.translators))
	for locale := range m.translators {
		locales = append(locales, locale)
	}
	return locales
}

// createTranslator 创建翻译器（带锁）
func (m *Manager) createTranslator(locale string) (Translator, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if translator, exists := m.translators[locale]; exists {
		return translator, nil
	}

	translator := m.getOrCreateTranslator(locale)
	return translator, nil
}

// getOrCreateTranslator 获取或创建翻译器（不加锁）
func (m *Manager) getOrCreateTranslator(locale string) Translator {
	if translator, exists := m.translators[locale]; exists {
		return translator
	}

	config := &TranslatorConfig{
		CurrentLocale:  locale,
		DefaultLocale:  m.config.DefaultLocale,
		FallbackLocale: m.config.FallbackLocale,
		MissingHandler: m.config.MissingHandler,
	}

	translator := NewTranslator(config)
	m.translators[locale] = translator
	// translator.SetLocale(locale)

	return translator
}

// 全局管理器实例
var (
	globalManager *Manager
	globalOnce    sync.Once
)

// InitGlobal 初始化全局管理器
func InitGlobal(config *ManagerConfig) {
	globalOnce.Do(func() {
		globalManager = NewManager(config)
	})
}

// GetGlobalManager 获取全局管理器
func GetGlobalManager() *Manager {
	if globalManager == nil {
		InitGlobal(nil)
	}
	return globalManager
}

// T 全局翻译函数
func T(key string, args ...interface{}) string {
	return GetGlobalManager().GetDefaultTranslator().T(key, args...)
}

// Tf 全局格式化翻译函数
func Tf(key string, data interface{}) string {
	return GetGlobalManager().GetDefaultTranslator().Tf(key, data)
}

// TWithLocale 指定语言环境的翻译函数
func TWithLocale(locale, key string, args ...interface{}) string {
	manager := GetGlobalManager()
	translator, err := manager.GetTranslator(locale)
	if err != nil {
		return key
	}
	return translator.T(key, args...)
}

// SetGlobalLocale 设置全局语言环境
func SetGlobalLocale(locale string) error {
	return GetGlobalManager().GetDefaultTranslator().SetLocale(locale)
}

// Context 相关的函数

type contextKey string

const localeContextKey contextKey = "i18n_locale"

// WithLocale 在 context 中设置语言环境
func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeContextKey, locale)
}

// GetLocaleFromContext 从 context 中获取语言环境
func GetLocaleFromContext(ctx context.Context) string {
	if locale, ok := ctx.Value(localeContextKey).(string); ok {
		return locale
	}
	return GetGlobalManager().config.DefaultLocale
}

// TWithContext 使用 context 中的语言环境进行翻译
func TWithContext(ctx context.Context, key string, args ...interface{}) string {
	locale := GetLocaleFromContext(ctx)
	return TWithLocale(locale, key, args...)
}

// TfWithContext 使用 context 中的语言环境进行格式化翻译
func TfWithContext(ctx context.Context, key string, data interface{}) string {
	locale := GetLocaleFromContext(ctx)
	translator, err := GetGlobalManager().GetTranslator(locale)
	if err != nil {
		return key
	}
	return translator.Tf(key, data)
}
