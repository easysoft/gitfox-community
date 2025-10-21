# I18n 国际化包快速入门指南

快速上手使用 GitFox Community 的国际化（i18n）包，实现多语言支持。

## 目录

- [基础概念](#基础概念)
- [从文件中加载多语言](#从文件中加载多语言)
- [使用多层调用多语言](#使用多层调用多语言)
- [使用模板占位符](#使用模板占位符)
- [外部项目集成](#外部项目集成)
- [多语言Key校验](#多语言key校验)
- [Web中间件集成](#web中间件集成)
- [最佳实践](#最佳实践)

## 基础概念

### 核心组件

- **Manager**: 国际化管理器，负责管理多个翻译器
- **Translator**: 翻译器，负责具体的翻译工作
- **Loader**: 文件加载器，负责从文件系统加载翻译文件
- **Extractor**: 字符串提取器，用于从代码中提取需要翻译的字符串

## 翻译文件格式

### YAML 格式（推荐）

本包支持 YAML 格式的翻译文件，支持嵌套结构并自动扁平化为点分隔的键：

```yaml
# locales/zh-CN/common.yaml
common:
  welcome: "欢迎"
  goodbye: "再见"
  
user:
  login_success: "登录成功"
  logout_success: "退出成功"
```

上述 YAML 文件会被解析为以下键值对：
- `common.welcome` → "欢迎"
- `common.goodbye` → "再见"
- `user.login_success` → "登录成功"
- `user.logout_success` → "退出成功"

### JSON 格式（兼容）

同样支持 JSON 格式的翻译文件：

```json
{
  "common": {
    "welcome": "欢迎",
    "goodbye": "再见"
  },
  "user": {
    "login_success": "登录成功",
    "logout_success": "退出成功"
  }
}
```

## 从文件中加载多语言

### 1. 准备翻译文件

创建多语言文件目录结构：
```
locales/
├── zh-CN/
│   ├── common.yaml
│   └── user.yaml
└── en-US/
    ├── common.yaml
    └── user.yaml
```

### 2. 编写翻译文件

**locales/zh-CN/space.yaml**:
```yaml
# 空间相关翻译
space:
  secrets_add_error: "添加密钥失败，请检查名称是否重复或格式是否正确。"
  secrets_add_success: "密钥已成功添加。"
  secrets_list_error: "无法加载密钥列表，请稍后重试。"
  secrets_list_success: "仓库密钥列表成功。"
  secrets_delete_used: "删除失败：此密钥可能正在被一个或多个服务使用。"
  secrets_delete_permission: "删除失败：权限不足，您无法删除此密钥。"
  secrets_delete_success: "密钥已成功删除。"
  artifacts_title_success: "仓库设置保存成功"
```

**locales/en-US/space.yaml**:
```yaml
# Space related translations
space:
  secrets_add_error: "Failed to add secret. Please check if the name is duplicated or the format is correct."
  secrets_add_success: "Secret added successfully."
  secrets_list_error: "Failed to load the list of secrets. Please try again later."
  secrets_list_success: "Repository Secrets List Loaded Successfully."
  secrets_delete_used: "Failed to delete: This secret may be in use by one or more services."
  secrets_delete_permission: "Failed to delete: You do not have permission to delete this secret."
  secrets_delete_success: "Secret deleted successfully."
  artifacts_title_success: "Repository Settings Saved Successfully."
```

### 3. 初始化管理器并加载文件

```go
package main

import (
    "fmt"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
    // 创建管理器配置
    config := &i18n.ManagerConfig{
        DefaultLocale:   "zh-CN",
        FallbackLocale:  "en-US", 
        TranslationsDir: "locales",
        AutoLoad:        true,
    }

    // 创建管理器
    manager := i18n.NewManager(config)

    // 从目录加载所有翻译文件
    err := manager.LoadTranslationsFromDir("locales")
    if err != nil {
        panic(err)
    }

    // 初始化全局管理器（可选）
    i18n.InitGlobal(config)

    fmt.Println("多语言文件加载完成！")
}
```

## 使用多层调用多语言

### 1. 基础翻译调用

```go
// 使用默认语言环境嵌套键翻译
loginText := i18n.T("space.secrets_add_success")
fmt.Println(loginText) // 输出: 密钥已成功添加。
```

### 2. 指定语言环境翻译

```go
// 指定语言环境翻译
welcomeEn := i18n.TWithLocale("en-US", "space.secrets_add_success")
fmt.Println(welcomeEn) // 输出: Secret added successfully.

welcomeZh := i18n.TWithLocale("zh-CN", "space.secrets_add_success")
fmt.Println(welcomeZh) // 输出: 密钥已成功添加。
```

### 3. 上下文相关翻译

```go
import "context"

func handleRequest(ctx context.Context) {
    // 从上下文中获取语言环境并翻译
    text := i18n.TWithContext(ctx, "welcome")
    fmt.Println(text)
}

// 设置上下文语言环境
ctx := i18n.WithLocale(context.Background(), "en-US")
handleRequest(ctx)
```

### 4. 获取翻译器实例

```go
// 获取特定语言的翻译器
translator, err := manager.GetTranslator("zh-CN")
if err != nil {
    panic(err)
}

// 使用翻译器
text := translator.T("space.secrets_add_success")
fmt.Println(text) // 输出: 密钥已成功添加。

// 切换语言环境
translator.SetLocale("en-US")
text = translator.T("space.secrets_add_success")
fmt.Println(text) // 输出: Secret added successfully.
```

## 使用模板占位符

### 1. 带参数的翻译

**locales/zh-CN/user.yaml**:
```yaml
# 用户相关翻译
user:
  greeting: "你好，%s！"
  user_count: "共有 %d 个用户"
  user_info: "用户：{{.Name}}，年龄：{{.Age}}"
```

**locales/en-US/user.yaml**:
```yaml
# User related translations
user:
  greeting: "Hello, %s!"
  user_count: "There are %d users"
  user_info: "User: {{.Name}}, Age: {{.Age}}"
```

### 2. 使用参数翻译

```go
// 使用 fmt.Sprintf 风格的参数
greeting := i18n.T("user.greeting", "张三")
fmt.Println(greeting) // 输出: 你好，张三！

userCount := i18n.T("user.user_count", 100)
fmt.Println(userCount) // 输出: 共有 100 个用户
```

### 3. 使用模板数据

```go
// 使用模板数据
userData := map[string]interface{}{
    "Name": "李四",
    "Age":  25,
}

userInfo := i18n.Tf("user.user_info", userData)
fmt.Println(userInfo) // 输出: 用户：李四，年龄：25

// 或者使用结构体
type User struct {
    Name string
    Age  int
}

user := User{Name: "王五", Age: 30}
userInfo = i18n.Tf("user.user_info", user)
fmt.Println(userInfo) // 输出: 用户：王五，年龄：30
```

### 4. 上下文模板翻译

```go
func handleUserInfo(ctx context.Context, user User) {
    userInfo := i18n.TfWithContext(ctx, "user.user_info", user)
    fmt.Println(userInfo)
}

// 使用
ctx := i18n.WithLocale(context.Background(), "en-US")
user := User{Name: "John", Age: 28}
handleUserInfo(ctx, user) // 输出: User: John, Age: 28
```

## 外部项目集成

### 1. 获取开源版翻译数据

外部项目可以使用 `GetProjectTranslations` 函数获取 gitfox-community 项目内置的翻译数据：

```go
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
    
    // 按语言获取翻译
    zhTranslations, err := i18n.GetProjectTranslationsByLocale("zh-CN")
    if err != nil {
        log.Fatal("获取中文翻译失败:", err)
    }
    
    enTranslations, err := i18n.GetProjectTranslationsByLocale("en-US")
    if err != nil {
        log.Fatal("获取英文翻译失败:", err)
    }
    
    fmt.Printf("中文翻译数量: %d\n", len(zhTranslations))
    fmt.Printf("英文翻译数量: %d\n", len(enTranslations))
}
```

### 2. 混合使用开源版翻译和自定义翻译

外部项目可以同时使用开源版的翻译和自己项目的翻译：

```go
package main

import (
    "log"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
    // 创建自己的翻译管理器
    manager := i18n.NewManager(&i18n.ManagerConfig{
        DefaultLocale:  "zh-CN",
        FallbackLocale: "en-US",
    })
    
    // 1. 加载开源版的翻译数据
    entries, err := i18n.GetProjectTranslations()
    if err != nil {
        log.Fatal("获取开源版翻译失败:", err)
    }
    
    // 将开源版翻译添加到管理器中
    for _, entry := range entries {
        manager.AddTranslation(entry.Locale, entry.Key, entry.Value)
    }
    
    // 2. 加载自己项目的翻译文件
    // 假设你的项目有自己的 locales 目录
    err = manager.LoadTranslationsFromDir("./locales")
    if err != nil {
        log.Fatal("加载自定义翻译失败:", err)
    }
    
    // 3. 手动添加特定的翻译
    manager.AddTranslation("zh-CN", "my_project.welcome", "欢迎使用我的项目")
    manager.AddTranslation("en-US", "my_project.welcome", "Welcome to my project")
    
    // 4. 使用翻译
    translator, _ := manager.GetTranslator("zh-CN")
    
    // 使用开源版的翻译
    fmt.Println(translator.T("space.secrets_add_success")) // 密钥已成功添加。
    
    // 使用自定义的翻译
    fmt.Println(translator.T("my_project.welcome")) // 欢迎使用我的项目
}
```

### 3. 作为依赖包使用

在你的项目中添加依赖：

```go
// go.mod
module your-project

require (
    gitfox.zcorp.cc/git/Pangu/gitfox-community v0.1.0
)
```

### 2. 初始化配置

```go
package main

import (
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func init() {
    // 初始化全局i18n管理器
    config := &i18n.ManagerConfig{
        DefaultLocale:   "zh-CN",
        FallbackLocale:  "en-US",
        TranslationsDir: "./locales",
        AutoLoad:        true,
    }
    
    i18n.InitGlobal(config)
    
    // 加载翻译文件
    manager := i18n.GetGlobalManager()
    if err := manager.LoadTranslationsFromDir("./locales"); err != nil {
        panic(err)
    }
}
```

### 3. 在业务代码中使用

```go
package service

import (
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

type UserService struct{}

func (s *UserService) CreateUser(name string) error {
    // 业务逻辑...
    
    if name == "" {
        // 返回国际化错误消息
        return fmt.Errorf(i18n.T("user.name_required"))
    }
    
    // 成功消息
    fmt.Println(i18n.T("user.create_success", name))
    return nil
}
```

### 4. 动态添加翻译

```go
// 运行时动态添加翻译
manager := i18n.GetGlobalManager()
manager.AddTranslation("zh-CN", "dynamic.message", "这是动态添加的消息")
manager.AddTranslation("en-US", "dynamic.message", "This is a dynamically added message")

// 使用动态添加的翻译
text := i18n.T("dynamic.message")
fmt.Println(text)
```

## 多语言Key校验

### 1. 使用内置校验工具

项目提供了 `i18n-check` 工具来校验多语言文件的键一致性：

```bash
# 编译校验工具
go build -o i18n-check ./pkg/i18n/i18n-check

# 运行校验
./i18n-check locales

# 输出示例：
# 🔍 检查多语言文件键一致性: locales
# ✅ 键一致性检查通过！
# 📊 检查的语言环境: [en-US zh-CN]
```

### 2. 编程方式校验

```go
package main

import (
    "fmt"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func validateTranslations() {
    manager := i18n.NewManager(&i18n.ManagerConfig{
        DefaultLocale:  "zh-CN",
        FallbackLocale: "en-US",
    })

    // LoadTranslationsFromDir 内部会自动进行键一致性校验
    err := manager.LoadTranslationsFromDir("locales")
    if err != nil {
        fmt.Printf("校验失败: %v\n", err)
        return
    }

    fmt.Println("所有语言文件键一致性校验通过！")
}
```

### 3. 自定义校验逻辑

```go
// 检查特定键是否存在
func checkKeyExists(manager *i18n.Manager, key string) {
    locales := manager.GetAvailableLocales()
    
    for _, locale := range locales {
        translator, err := manager.GetTranslator(locale)
        if err != nil {
            continue
        }
        
        if !translator.HasTranslation(locale, key) {
            fmt.Printf("警告: 语言 %s 缺少键 %s\n", locale, key)
        }
    }
}
```

## Web中间件集成

### 1. Gin 框架集成

```go
package main

import (
    "github.com/gin-gonic/gin"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
    // 初始化i18n
    i18n.InitGlobal(&i18n.ManagerConfig{
        DefaultLocale:  "zh-CN",
        FallbackLocale: "en-US",
    })

    r := gin.Default()

    // 添加i18n中间件
    middlewareConfig := &i18n.MiddlewareConfig{
        DefaultLocale:   "zh-CN",
        AcceptedLocales: []string{"zh-CN", "en-US"},
        // QueryParam:      "lang",     // ?lang=en-US
        // CookieName:      "locale",   // Cookie中的语言设置
        HeaderName:      "Accept-Language", // HTTP头中的语言
    }
    
    r.Use(i18n.GinMiddleware(middlewareConfig))

    r.GET("/hello", func(c *gin.Context) {
        // 使用Gin上下文获取翻译
        message := i18n.TGin(c, "welcome")
        c.JSON(200, gin.H{
            "message": message,
            "locale":  i18n.GetLocaleFromGin(c),
        })
    })

    r.Run(":8080")
}
```

### 2. 标准HTTP中间件

```go
package main

import (
    "net/http"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
    // 初始化i18n
    i18n.InitGlobal(&i18n.ManagerConfig{
        DefaultLocale:  "zh-CN", 
        FallbackLocale: "en-US",
    })

    mux := http.NewServeMux()

    // 添加i18n中间件
    middlewareConfig := &i18n.MiddlewareConfig{
        DefaultLocale:   "zh-CN",
        AcceptedLocales: []string{"zh-CN", "en-US"},
        QueryParam:      "lang",
        HeaderName:      "Accept-Language",
    }

    handler := i18n.HTTPMiddleware(middlewareConfig)(mux)

    mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        // 从请求上下文获取翻译
        message := i18n.TWithContext(r.Context(), "welcome")
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message":"` + message + `"}`))
    })

    http.ListenAndServe(":8080", handler)
}
```

## 总结

通过本快速入门指南，你已经学会了：

1. ✅ 从文件中加载多语言配置
2. ✅ 使用多层调用进行翻译
3. ✅ 使用模板占位符处理动态内容
4. ✅ 在外部项目中集成i18n包
5. ✅ 校验多语言文件的键一致性
6. ✅ 在Web框架中使用中间件
7. ✅ 遵循最佳实践

现在你可以在项目中轻松实现国际化支持了！如有问题，请参考完整的API文档或提交Issue。