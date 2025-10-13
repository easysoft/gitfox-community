好的，已为您修复并优化了 README 文档的格式。

# I18n Package

一个功能完整的 Go 语言国际化（i18n）包，提供字符串翻译、多语言支持、中间件集成等功能。

## 特性

- 🌍 多语言支持
- 🔄 动态语言切换
- 📝 模板翻译支持
- 🚀 高性能并发安全
- 🔌 Gin/HTTP 中间件集成
- 📁 JSON 文件格式支持
- 🛠️ 字符串提取工具
- ✅ 翻译验证工具
- 🎯 Context 支持
- 🔧 灵活的配置选项
- 📂 嵌套JSON结构支持
- 🔍 自动语言环境检测

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
    // 初始化全局管理器
    i18n.InitGlobal(&i18n.ManagerConfig{
        DefaultLocale:   "zh-CN",      // 默认语言环境
        FallbackLocale:  "en-US",      // 回退语言环境
        TranslationsDir: "locales", // 翻译文件目录
        AutoLoad:        true,      // 自动加载翻译文件
    })

    // 添加翻译
    manager := i18n.GetGlobalManager()
    manager.AddTranslation("en-US", "hello", "Hello")
    manager.AddTranslation("zh-CN", "hello", "你好")
    manager.AddTranslation("en-US", "welcome", "Welcome, %s!")
    manager.AddTranslation("zh-CN", "welcome", "欢迎，%s！")

    // 使用翻译
    fmt.Println(i18n.T("hello"))                      // Hello (使用默认语言环境)
    fmt.Println(i18n.TWithLocale("zh-CN", "hello"))   // 你好
    fmt.Println(i18n.T("welcome", "Alice"))           // Welcome, Alice!
}
```

### 模板翻译

```go
// 添加模板翻译
manager.AddTranslation("en-US", "user_info", "User: {{.Name}}, Age: {{.Age}}")
manager.AddTranslation("zh-CN", "user_info", "用户：{{.Name}}，年龄：{{.Age}}")

// 使用模板翻译
data := map[string]interface{}{
    "Name": "Alice",
    "Age":  25,
}
fmt.Println(i18n.Tf("user_info", data)) // User: Alice, Age: 25
```

### Context 支持

```go
import "context"

// 在 context 中设置语言环境
ctx := i18n.WithLocale(context.Background(), "zh-CN")

// 使用 context 中的语言环境
fmt.Println(i18n.TWithContext(ctx, "hello"))        // 你好
fmt.Println(i18n.TfWithContext(ctx, "user_info", data)) // 用户：Alice，年龄：25
```

### Gin 中间件

```go
import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()

    // 使用 i18n 中间件
    r.Use(i18n.GinMiddleware(&i18n.MiddlewareConfig{
        DefaultLocale:   "en-US",
        AcceptedLocales: []string{"en-US", "zh-CN", "ja"},
        QueryParam:      "lang",     // 从查询参数获取语言
        CookieName:      "locale",   // 从Cookie获取语言
        HeaderName:      "X-Locale", // 从HTTP头获取语言
        SessionKey:      "locale",   // 从Session获取语言
    }))

    r.GET("/api/hello", func(c *gin.Context) {
        message := i18n.TGin(c, "hello")
        c.JSON(200, gin.H{"message": message})
    })

    r.Run(":8080")
}
```

### HTTP 中间件

```go
import "net/http"

func main() {
    mux := http.NewServeMux()

    // 使用 HTTP 中间件
    handler := i18n.HTTPMiddleware(&i18n.MiddlewareConfig{
        DefaultLocale:   "en-US",
        AcceptedLocales: []string{"en-US", "zh-CN"},
        QueryParam:      "lang",
        HeaderName:      "Accept-Language",
    })(mux)

    http.ListenAndServe(":8080", handler)
}
```

## 翻译文件格式

### 嵌套 JSON 格式

本包支持嵌套的JSON结构，可以更好地组织翻译内容：

```json
{
  "space": {
    "profile": {
      "title": {
        "description": "Code repository user profile page title",
        "other": "Repository Profile"
      },
      "saveButton": {
        "description": "Code repository profile page save button",
        "other": "Save Changes"
      }
    },
    "settings": {
      "title": {
        "description": "Code repository user settings page title",
        "other": "Repository Settings"
      }
    }
  }
}
```

使用嵌套翻译：

```go
// 访问嵌套的翻译键
fmt.Println(i18n.T("space.profile.title"))      // Repository Profile
fmt.Println(i18n.T("space.profile.saveButton")) // Save Changes
```

### 简单 JSON 格式

也支持传统的扁平化JSON格式：

```json
{
  "hello": "Hello",
  "welcome": "Welcome, %s!",
  "user_info": "User: {{.Name}}, Age: {{.Age}}"
}
```

### 文件命名规范

支持以下文件命名格式：

-   `en_US.json` - 美式英文翻译
-   `zh_CN.json` - 简体中文翻译
-   `zh_TW.json` - 繁体中文翻译
-   `messages.en_US.json` - 带前缀的美式英文翻译
-   `app.zh_CN.json` - 带前缀的简体中文翻译

## 目录结构

推荐的翻译文件目录结构：

```
locales/
├── en_US/
│   ├── space.json
│   └── repo.json
└── zh_CN/
    ├── space.json
    └── repo.json
```

## API 参考

### 全局函数

-   `InitGlobal(config *ManagerConfig)` - 初始化全局管理器
-   `GetGlobalManager() *Manager` - 获取全局管理器
-   `T(key string, args ...interface{}) string` - 全局翻译函数
-   `Tf(key string, data interface{}) string` - 全局格式化翻译函数
-   `TWithLocale(locale, key string, args ...interface{}) string` - 指定语言环境的翻译函数
-   `SetGlobalLocale(locale string) error` - 设置全局语言环境

### Context 函数

-   `WithLocale(ctx context.Context, locale string) context.Context` - 在context中设置语言环境
-   `GetLocaleFromContext(ctx context.Context) string` - 从context中获取语言环境
-   `TWithContext(ctx context.Context, key string, args ...interface{}) string` - 使用context中的语言环境进行翻译
-   `TfWithContext(ctx context.Context, key string, data interface{}) string` - 使用context中的语言环境进行格式化翻译

### Gin 辅助函数

-   `GetLocaleFromGin(c *gin.Context) string` - 从Gin context获取语言环境
-   `TGin(c *gin.Context, key string, args ...interface{}) string` - Gin翻译函数
-   `TfGin(c *gin.Context, key string, data interface{}) string` - Gin格式化翻译函数

### 管理器配置

```go
type ManagerConfig struct {
    DefaultLocale   string                    // 默认语言环境
    FallbackLocale  string                    // 回退语言环境
    TranslationsDir string                    // 翻译文件目录
    AutoLoad        bool                      // 是否自动加载翻译文件
    MissingHandler  MissingTranslationHandler // 缺失翻译处理器
}
```

### 中间件配置

```go
type MiddlewareConfig struct {
    DefaultLocale   string   // 默认语言环境
    AcceptedLocales []string // 接受的语言环境列表
    QueryParam      string   // 查询参数名
    CookieName      string   // Cookie名称
    HeaderName      string   // HTTP头名称
    SessionKey      string   // Session键名
}
```

## 工具函数

### 翻译验证

```go
// 验证翻译完整性
result, err := i18n.ValidateTranslations("locales/en_US.json", []string{
    "locales/zh_CN.json",
    "locales/ja.json",
})

if err != nil {
    log.Fatal(err)
}

// 检查验证结果
for file, fileResult := range result.FileResults {
    fmt.Printf("File: %s, Coverage: %.2f%%\n", file, fileResult.Coverage)
    if len(fileResult.MissingKeys) > 0 {
        fmt.Printf("Missing keys: %v\n", fileResult.MissingKeys)
    }
}
```

### 字符串提取

```go
// 创建字符串提取器
extractor := i18n.NewExtractor(&i18n.ExtractorConfig{
    FunctionNames:   []string{"T", "Tf", "TGin", "TfGin"},
    MinLength:       2,
    MaxLength:       100,
    IncludeComments: false,
})

// 从Go文件中提取字符串
strings, err := extractor.ExtractFromFile("main.go")
if err != nil {
    log.Fatal(err)
}
```

## 支持的语言环境

包内置支持以下语言环境：

-   `en` - English
-   `zh-CN` - Simplified Chinese
-   `zh-TW` - Traditional Chinese
-   `ja` - Japanese
-   `ko` - Korean
-   `fr` - French
-   `de` - German
-   `es` - Spanish
-   `pt` - Portuguese
-   `ru` - Russian

## 最佳实践

1.  **使用嵌套结构组织翻译**：将相关的翻译键组织在一起，便于管理
2.  **提供描述信息**：在翻译对象中添加`description`字段，帮助翻译人员理解上下文
3.  **设置回退语言**：确保在缺失翻译时有合适的回退选项
4.  **使用语言环境检测**：在Web应用中使用中间件自动检测用户的语言偏好
5.  **验证翻译完整性**：定期使用验证工具检查翻译文件的完整性
6.  **合理的目录结构**：按语言环境和功能模块组织翻译文件

## Demo

查看 `pkg/demo/i18n/main.go` 获取完整的使用示例。