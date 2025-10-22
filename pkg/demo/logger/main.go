package main

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. 基础同步模式
	syncLog := logger.NewZapMateLogger("./demo_log.yaml", "gitfox")
	syncLog.Info("同步日志消息", zap.String("mode", "sync"))

	// 2. 异步模式
	asyncLog := logger.NewZapMateLogger("./demo_log.yaml", "gitfox").SetAsyncer(1000).Sugar()
	asyncLog.Info("异步日志消息", zap.String("mode", "async"))
	asyncLog.Flush() // 确保异步日志写入完成

	// 3. 命名 Logger
	namedLog := asyncLog.Named("user-service")
	namedLog.Info("用户服务日志", zap.String("action", "login"))

	// 4. 带字段的 Logger
	logWithFields := namedLog.With(
		zap.String("request_id", "req-123"),
		zap.String("user_id", "user-456"),
	)
	logWithFields.Info("处理用户请求1")
	logWithFields.Info("处理用户请求2")
	logWithFields.Info("处理用户请求3")
	logWithFields.Flush()

	// 5. Sugar 模式 - 简单日志
	sugar := asyncLog.Sugar()
	sugar.Info("Sugar 简单日志")

	// 6. Sugar 模式 - 格式化日志
	sugar.Infof("用户 %s 在 %v 登录", "张三", time.Now().Format("2006-01-02 15:04:05"))

	// 7. Sugar 模式 - 键值对日志
	sugar.Infow("用户操作详情",
		"action", "purchase",
		"user_id", "user-789",
		"product_id", "prod-101",
		"amount", 99.99,
		"timestamp", time.Now(),
	)

	// 8. Sugar 模式 - 带命名和字段
	namedSugar := sugar.Named("payment-service").With("service", "payment", "version", "v1.0")
	namedSugar.Infof("用户 %s 在 %v 准备支付", "张三", time.Now().Format("2006-01-02 15:04:05"))
	namedSugar.Infow("支付处理",
		"amount", 199.99,
	)
	namedSugar.Infow("支付结果",
		"status", "success",
	)

	// 9. 不同日志级别示例
	sugar.Debug("调试信息 - 仅在开发环境显示")
	sugar.Info("一般信息 - 正常运行信息")
	sugar.Warn("警告信息 - 需要注意但不影响运行")
	sugar.Error("错误信息 - 发生了错误但程序可以继续")

	// 10. 错误处理示例
	if err := someOperation(); err != nil {
		sugar.Errorw("操作失败",
			"error", err,
			"operation", "someOperation",
			"retry_count", 3,
		)
	}

	// 确保所有日志都写入
	sugar.Flush()
}

func someOperation() error {
	// 模拟一个可能出错的操作
	return nil
}
