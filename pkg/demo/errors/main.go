package main

import (
	"fmt"
	"os"

	pkg_errors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

func main() {
	fmt.Println("=== GitFox Errors Demo ===\n")

	// 1. 基本错误创建
	fmt.Println("1. 基本错误创建:")
	basicErr := pkg_errors.New("这是一个基本错误")
	fmt.Printf("   基本错误: %v\n", basicErr)
	fmt.Printf("   错误类型: %v\n\n", pkg_errors.GetType(basicErr))

	// 2. 格式化错误创建
	fmt.Println("2. 格式化错误创建:")
	formattedErr := pkg_errors.Newf("用户 %s 不存在，ID: %d", "张三", 123)
	fmt.Printf("   格式化错误: %v\n", formattedErr)
	fmt.Printf("   错误类型: %v\n\n", pkg_errors.GetType(formattedErr))

	// 3. 类型化错误创建
	fmt.Println("3. 类型化错误创建:")
	badRequestErr := pkg_errors.BadRequest.New("请求参数无效")
	notFoundErr := pkg_errors.NotFound.New("资源未找到")
	dbErr := pkg_errors.DBError.New("数据库连接失败")
	systemErr := pkg_errors.SystemErr.New("系统内部错误")

	fmt.Printf("   BadRequest 错误: %v (类型: %v)\n", badRequestErr, pkg_errors.GetType(badRequestErr))
	fmt.Printf("   NotFound 错误: %v (类型: %v)\n", notFoundErr, pkg_errors.GetType(notFoundErr))
	fmt.Printf("   DBError 错误: %v (类型: %v)\n", dbErr, pkg_errors.GetType(dbErr))
	fmt.Printf("   SystemErr 错误: %v (类型: %v)\n\n", systemErr, pkg_errors.GetType(systemErr))

	// 4. 错误包装
	fmt.Println("4. 错误包装:")
	originalErr := fmt.Errorf("原始错误: 网络连接超时")
	wrappedErr := pkg_errors.Wrap(originalErr, "处理用户请求时发生错误")
	fmt.Printf("   原始错误: %v\n", originalErr)
	fmt.Printf("   包装错误: %v\n", wrappedErr)
	fmt.Printf("   获取原因: %v\n\n", pkg_errors.Cause(wrappedErr))

	// 5. 类型化错误包装
	fmt.Println("5. 类型化错误包装:")
	networkErr := fmt.Errorf("网络不可达")
	typedWrappedErr := pkg_errors.DBError.Wrap(networkErr, "数据库连接失败")
	fmt.Printf("   网络错误: %v\n", networkErr)
	fmt.Printf("   类型化包装: %v (类型: %v)\n", typedWrappedErr, pkg_errors.GetType(typedWrappedErr))
	fmt.Printf("   获取原因: %v\n\n", pkg_errors.Cause(typedWrappedErr))

	// 6. 格式化包装
	fmt.Println("6. 格式化包装:")
	baseErr := fmt.Errorf("权限不足")
	formattedWrapErr := pkg_errors.BadRequest.Wrapf(baseErr, "用户 %s 尝试访问资源 %s", "李四", "/admin/users")
	fmt.Printf("   基础错误: %v\n", baseErr)
	fmt.Printf("   格式化包装: %v\n", formattedWrapErr)
	fmt.Printf("   错误类型: %v\n\n", pkg_errors.GetType(formattedWrapErr))

	// 7. 错误消息添加
	fmt.Println("7. 错误消息添加:")
	simpleErr := pkg_errors.New("文件不存在")
	withMsgErr := pkg_errors.WithMessage(simpleErr, "读取配置文件时")
	withMsgfErr := pkg_errors.WithMessagef(simpleErr, "在 %s 目录下读取 %s 文件时", "/etc/config", "app.yaml")
	fmt.Printf("   简单错误: %v\n", simpleErr)
	fmt.Printf("   添加消息: %v\n", withMsgErr)
	fmt.Printf("   格式化消息: %v\n\n", withMsgfErr)

	// 8. 错误上下文
	fmt.Println("8. 错误上下文:")
	contextErr := pkg_errors.New("验证失败")
	contextErr = pkg_errors.AddErrorContext(contextErr, "username", "用户名不能为空")
	contextErr = pkg_errors.AddErrorContext(contextErr, "password", "密码长度不足")

	fmt.Printf("   带上下文错误: %v\n", contextErr)
	context := pkg_errors.GetErrorContext(contextErr)
	fmt.Printf("   错误上下文: %+v\n\n", context)

	// 9. 错误判断
	fmt.Println("9. 错误判断:")
	notFoundError := pkg_errors.NotFound.New("用户不存在")
	fmt.Printf("   NotFound 错误: %v\n", notFoundError)
	fmt.Printf("   是否为 NotFound: %v\n", pkg_errors.IsNotFound(notFoundError))
	fmt.Printf("   是否为 BadRequest: %v\n", pkg_errors.GetType(notFoundError) == pkg_errors.BadRequest)

	// 错误比较
	targetErr := fmt.Errorf("目标错误")
	wrappedTargetErr := pkg_errors.Wrap(targetErr, "包装的目标错误")
	fmt.Printf("   目标错误: %v\n", targetErr)
	fmt.Printf("   包装错误: %v\n", wrappedTargetErr)
	fmt.Printf("   错误匹配: %v\n\n", pkg_errors.Is(wrappedTargetErr, targetErr))

	// 10. 实际业务场景示例
	fmt.Println("10. 实际业务场景示例:")

	// 用户登录场景
	if err := simulateUserLogin("", "123456"); err != nil {
		fmt.Printf("   登录失败: %v\n", err)
		fmt.Printf("   错误类型: %v\n", pkg_errors.GetType(err))
		if context := pkg_errors.GetErrorContext(err); context != nil {
			fmt.Printf("   错误详情: %+v\n", context)
		}
	}

	// 数据库操作场景
	if err := simulateDBOperation(); err != nil {
		fmt.Printf("   数据库操作失败: %v\n", err)
		fmt.Printf("   错误类型: %v\n", pkg_errors.GetType(err))
		fmt.Printf("   根本原因: %v\n", pkg_errors.Cause(err))
	}

	// 文件操作场景
	if err := simulateFileOperation(); err != nil {
		fmt.Printf("   文件操作失败: %v\n", err)
		fmt.Printf("   错误类型: %v\n", pkg_errors.GetType(err))
		if pkg_errors.IsNotFound(err) {
			fmt.Printf("   这是一个 NotFound 错误\n")
		}
	}

	fmt.Println("\n=== Demo 完成 ===")
}

// 模拟用户登录
func simulateUserLogin(username, password string) error {
	if username == "" {
		err := pkg_errors.BadRequest.New("用户名不能为空")
		return pkg_errors.AddErrorContext(err, "field", "username")
	}

	if len(password) < 6 {
		err := pkg_errors.BadRequest.New("密码长度不足")
		return pkg_errors.AddErrorContext(err, "field", "password")
	}

	return nil
}

// 模拟数据库操作
func simulateDBOperation() error {
	// 模拟网络错误
	networkErr := fmt.Errorf("connection timeout")

	// 包装为数据库错误
	dbErr := pkg_errors.DBError.Wrap(networkErr, "连接数据库失败")

	// 再次包装为业务错误
	return pkg_errors.Wrapf(dbErr, "查询用户信息失败")
}

// 模拟文件操作
func simulateFileOperation() error {
	// 模拟文件不存在
	if _, err := os.Stat("nonexistent.txt"); os.IsNotExist(err) {
		return pkg_errors.NotFound.Wrap(err, "配置文件不存在")
	}

	return nil
}
