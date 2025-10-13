package main

import (
	"context"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

func main() {
	fmt.Println("=== GitFox I18n Demo ===\n")

	// 1. 基本翻译器创建和使用
	fmt.Println("1. 基本翻译器创建和使用:")
	basicTranslatorDemo()

	// 2. 管理器模式
	fmt.Println("\n2. 管理器模式:")
	managerDemo()

	// 3. 从文件加载翻译
	fmt.Println("\n3. 从文件加载翻译:")
	fileLoadingDemo()

	// 4. 全局翻译器
	fmt.Println("\n4. 全局翻译器:")
	globalTranslatorDemo()

	// 5. 上下文翻译
	fmt.Println("\n5. 上下文翻译:")
	contextTranslationDemo()

	// 6. 模板翻译
	fmt.Println("\n6. 模板翻译:")
	templateTranslationDemo()

	// 7. 缺失翻译处理
	fmt.Println("\n7. 缺失翻译处理:")
	missingTranslationDemo()

	// 8. 多语言切换
	fmt.Println("\n8. 多语言切换:")
	languageSwitchDemo()
}

func basicTranslatorDemo() {
	// 创建翻译器
	config := &i18n.TranslatorConfig{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
	}
	translator := i18n.NewTranslator(config)

	// 添加翻译
	translator.AddTranslation("zh-CN", "hello", "你好")
	translator.AddTranslation("zh-CN", "welcome", "欢迎 %s")
	translator.AddTranslation("zh-CN", "goodbye", "再见")

	translator.AddTranslation("en-US", "hello", "Hello")
	translator.AddTranslation("en-US", "welcome", "Welcome %s")
	translator.AddTranslation("en-US", "goodbye", "Goodbye")

	// 基本翻译
	fmt.Printf("   中文翻译: %s\n", translator.T("hello"))
	fmt.Printf("   格式化翻译: %s\n", translator.T("welcome", "张三"))

	// 切换语言
	translator.SetLocale("en_US")
	fmt.Printf("   英文翻译: %s\n", translator.T("hello"))
	fmt.Printf("   格式化翻译: %s\n", translator.T("welcome", "John"))

	// 显示可用语言
	fmt.Printf("   可用语言: %v\n", translator.GetAvailableLocales())
}

func managerDemo() {
	// 创建管理器
	config := &i18n.ManagerConfig{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
	}
	manager := i18n.NewManager(config)

	// 添加翻译
	manager.AddTranslation("zh-CN", "user.login", "用户登录")
	manager.AddTranslation("zh-CN", "user.logout", "用户退出")
	manager.AddTranslation("zh-CN", "user.profile", "用户资料")

	manager.AddTranslation("en-US", "user.login", "User Login")
	manager.AddTranslation("en-US", "user.logout", "User Logout")
	manager.AddTranslation("en-US", "user.profile", "User Profile")

	// 获取不同语言的翻译器
	zhTranslator, _ := manager.GetTranslator("zh-CN")
	enTranslator, _ := manager.GetTranslator("en-US")

	fmt.Printf("   中文翻译器: %s\n", zhTranslator.T("user.login"))
	fmt.Printf("   英文翻译器: %s\n", enTranslator.T("user.login"))

	// 显示所有可用语言
	fmt.Printf("   管理器可用语言: %v\n", manager.GetAvailableLocales())
}

func fileLoadingDemo() {
	// 创建管理器
	manager := i18n.NewManager(&i18n.ManagerConfig{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
	})

	// 尝试从真实的 locales 目录加载翻译
	localesDir := "../../../locales"
	err := manager.LoadTranslationsFromDir(localesDir)
	if err != nil {
		fmt.Printf("   加载翻译文件失败: %v\n", err)
		return
	} else {
		fmt.Printf("   成功从 %s 加载翻译文件\n", localesDir)
	}

	// 测试加载的翻译
	zhTranslator, _ := manager.GetTranslator("zh-CN")
	enTranslator, _ := manager.GetTranslator("en-US")

	fmt.Printf("   中文 - 密钥列表: %s\n", zhTranslator.T("space.secrets.list.success"))
	fmt.Printf("   中文 - 成功删除分支: %s\n", zhTranslator.T("repo.branches.delete.success"))
	fmt.Printf("   中文 - 删除密钥失败1: %s\n", zhTranslator.T("space.secrets.delete.error.error1"))
	fmt.Printf("   英文 - 密钥列表: %s\n", enTranslator.T("space.secrets.list.success"))
	fmt.Printf("   英文 - 成功删除分支: %s\n", enTranslator.T("repo.branches.delete.success"))
	fmt.Printf("   英文 - 删除密钥失败1: %s\n", enTranslator.T("space.secrets.delete.error.error1"))
}

func globalTranslatorDemo() {
	// 初始化全局翻译器
	i18n.InitGlobal(&i18n.ManagerConfig{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
	})

	// 获取全局管理器并添加翻译
	manager := i18n.GetGlobalManager()
	manager.AddTranslation("zh-CN", "app.name", "GitFox 社区版")
	manager.AddTranslation("zh-CN", "app.version", "版本 %s")
	manager.AddTranslation("zh-CN", "app.welcome", "欢迎使用 GitFox")

	manager.AddTranslation("en-US", "app.name", "GitFox Community")
	manager.AddTranslation("en-US", "app.version", "Version %s")
	manager.AddTranslation("en-US", "app.welcome", "Welcome to GitFox")

	// 使用全局翻译函数
	fmt.Printf("   全局翻译 - 应用名称: %s\n", i18n.T("app.name"))
	fmt.Printf("   全局翻译 - 应用版本: %s\n", i18n.T("app.version", "1.0.0"))
	fmt.Printf("   全局翻译 - 欢迎信息: %s\n", i18n.T("app.welcome"))

	// 切换全局语言
	i18n.SetGlobalLocale("en-US")
	fmt.Printf("   切换到英文后 - 应用名称: %s\n", i18n.T("app.name"))
	fmt.Printf("   切换到英文后 - 应用版本: %s\n", i18n.T("app.version", "1.0.0"))
}

func contextTranslationDemo() {
	// 初始化全局翻译器（如果还没有初始化）
	manager := i18n.GetGlobalManager()
	if manager == nil {
		i18n.InitGlobal(&i18n.ManagerConfig{
			DefaultLocale:  "zh-CN",
			FallbackLocale: "en-US",
		})
		manager = i18n.GetGlobalManager()
	}

	// 添加翻译
	manager.AddTranslation("zh-CN", "request.processing", "正在处理请求...")
	manager.AddTranslation("zh-CN", "request.completed", "请求处理完成")
	manager.AddTranslation("en-US", "request.processing", "Processing request...")
	manager.AddTranslation("en-US", "request.completed", "Request completed")

	// 创建带语言环境的上下文
	zhCtx := i18n.WithLocale(context.Background(), "zh-CN")
	enCtx := i18n.WithLocale(context.Background(), "en-US")

	// 使用上下文翻译
	fmt.Printf("   中文上下文: %s\n", i18n.TWithContext(zhCtx, "request.processing"))
	fmt.Printf("   英文上下文: %s\n", i18n.TWithContext(enCtx, "request.processing"))

	// 模拟处理请求
	simulateRequest(zhCtx, "用户登录")
	simulateRequest(enCtx, "User Login")
}

func templateTranslationDemo() {
	// 创建翻译器
	translator := i18n.NewTranslator(&i18n.TranslatorConfig{
		DefaultLocale:  "zh_CN",
		FallbackLocale: "en_US",
	})

	// 添加模板翻译
	translator.AddTranslation("zh_CN", "user.info", "用户: {{.Name}}, 年龄: {{.Age}}, 邮箱: {{.Email}}")
	translator.AddTranslation("zh_CN", "order.summary", "订单 #{{.OrderID}} - 总计: ¥{{.Total}}, 状态: {{.Status}}")

	translator.AddTranslation("en_US", "user.info", "User: {{.Name}}, Age: {{.Age}}, Email: {{.Email}}")
	translator.AddTranslation("en_US", "order.summary", "Order #{{.OrderID}} - Total: ${{.Total}}, Status: {{.Status}}")

	// 准备数据
	userData := map[string]interface{}{
		"Name":  "张三",
		"Age":   25,
		"Email": "zhangsan@example.com",
	}

	orderData := map[string]interface{}{
		"OrderID": "12345",
		"Total":   99.99,
		"Status":  "已完成",
	}

	// 使用模板翻译
	fmt.Printf("   中文模板 - 用户信息: %s\n", translator.Tf("user.info", userData))
	fmt.Printf("   中文模板 - 订单摘要: %s\n", translator.Tf("order.summary", orderData))

	// 切换到英文
	translator.SetLocale("en_US")
	userData["Name"] = "John"
	orderData["Status"] = "Completed"

	fmt.Printf("   英文模板 - 用户信息: %s\n", translator.Tf("user.info", userData))
	fmt.Printf("   英文模板 - 订单摘要: %s\n", translator.Tf("order.summary", orderData))
}

func missingTranslationDemo() {
	// 创建带自定义缺失处理器的翻译器
	config := &i18n.TranslatorConfig{
		DefaultLocale:  "zh_CN",
		FallbackLocale: "en_US",
		MissingHandler: func(locale, key string) string {
			return fmt.Sprintf("[缺失翻译: %s@%s]", key, locale)
		},
	}
	translator := i18n.NewTranslator(config)

	// 只添加部分翻译
	translator.AddTranslation("zh_CN", "existing.key", "存在的翻译")
	translator.AddTranslation("en_US", "existing.key", "Existing Translation")

	// 测试存在的翻译
	fmt.Printf("   存在的翻译: %s\n", translator.T("existing.key"))

	// 测试缺失的翻译
	fmt.Printf("   缺失的翻译: %s\n", translator.T("missing.key"))

	// 切换语言后测试
	translator.SetLocale("en_US")
	fmt.Printf("   英文缺失的翻译: %s\n", translator.T("missing.key"))
}

func languageSwitchDemo() {
	// 创建管理器
	manager := i18n.NewManager(&i18n.ManagerConfig{
		DefaultLocale:  "zh_CN",
		FallbackLocale: "en_US",
	})

	// 添加多语言翻译
	languages := map[string]map[string]string{
		"zh_CN": {
			"menu.home":     "首页",
			"menu.about":    "关于我们",
			"menu.contact":  "联系我们",
			"menu.settings": "设置",
		},
		"en_US": {
			"menu.home":     "Home",
			"menu.about":    "About Us",
			"menu.contact":  "Contact Us",
			"menu.settings": "Settings",
		},
		"ja_JP": {
			"menu.home":     "ホーム",
			"menu.about":    "私たちについて",
			"menu.contact":  "お問い合わせ",
			"menu.settings": "設定",
		},
	}

	// 加载所有翻译
	for locale, translations := range languages {
		for key, value := range translations {
			manager.AddTranslation(locale, key, value)
		}
	}

	// 演示语言切换
	locales := []string{"zh_CN", "en_US", "ja_JP"}
	menuKeys := []string{"menu.home", "menu.about", "menu.contact", "menu.settings"}

	for _, locale := range locales {
		fmt.Printf("   === %s 语言菜单 ===\n", locale)
		translator, err := manager.GetTranslator(locale)
		if err != nil {
			fmt.Printf("   获取翻译器失败: %v\n", err)
			continue
		}

		for _, key := range menuKeys {
			fmt.Printf("   %s: %s\n", key, translator.T(key))
		}
		fmt.Println()
	}

	fmt.Printf("   所有可用语言: %v\n", manager.GetAvailableLocales())
}

// 辅助函数：模拟请求处理
func simulateRequest(ctx context.Context, operation string) {
	fmt.Printf("   开始处理操作: %s\n", operation)
	fmt.Printf("   状态: %s\n", i18n.TWithContext(ctx, "request.processing"))

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("   状态: %s\n", i18n.TWithContext(ctx, "request.completed"))
}
