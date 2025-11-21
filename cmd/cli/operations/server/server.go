package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quicklyon/kingpin/v2"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/cron"
)

type command struct {
	mode string
}

func (c *command) run(*kingpin.ParseContext) error {
	app, cleanup, err := NewApp(c.mode)
	if err != nil {
		// 这里无法使用 app.SystemLog 因为 app 初始化失败，只能使用 fmt.Printf
		fmt.Printf("[Server] Failed to initialize app server: %v\n", err)
		panic(err)
	}
	defer func() {
		app.SystemLog.Info("[Server] Cleaning up resources...")
		cleanup()
		// Flush 日志确保所有异步日志都被写入
		_ = app.SystemLog.Flush()
	}()

	// 启动定时任务管理器
	cronManager := cron.NewCronManager()

	// 示例：添加系统健康检查定时任务：每一分钟打印一条系统日志
	_, err = cronManager.AddJob("system-health-check", "0 * * * * *", func() {
		app.SystemLog.Infow("System health check", "status", "running")
	})
	if err != nil {
		app.SystemLog.Errorw("[Server] Failed to add cron job", "error", err)
	} else {
		app.SystemLog.Info("[Server] System health check cron job started (runs every minute)")
	}

	// 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", app.SystemConfig.Server.Host, app.SystemConfig.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router.NewRouter(c.mode, app.APILog.SugaredMateLogger, app.Controller),
	}

	adminPAT, err := app.UserService.IssueAdminPAT()
	if err != nil {
		app.SystemLog.Errorw("[Server] Failed to issue admin PAT", "error", err)
		panic(err)
	}
	app.SystemLog.Infow("[Server] Admin PAT issued", "pat", adminPAT)

	// 创建错误通道用于捕获服务器启动错误
	serverErrors := make(chan error, 2) // 增加容量以支持 HTTP 和 SSH 服务器

	// 在 goroutine 中启动 HTTP 服务器
	go func() {
		app.SystemLog.Infow("[Server] Starting HTTP server", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.SystemLog.Errorw("[Server] HTTP server error", "error", err)
			serverErrors <- err
		}
	}()

	// 在 goroutine 中启动 SSH 服务器(如果配置启用)
	if app.SystemConfig.SSH.Enable {
		go func() {
			sshAddr := fmt.Sprintf("%s:%d", app.SystemConfig.SSH.Host, app.SystemConfig.SSH.Port)
			app.SystemLog.Infow("[Server] Starting SSH server", "address", sshAddr)
			if err := app.SSHServer.ListenAndServe(); err != nil {
				app.SystemLog.Errorw("[Server] SSH server error", "error", err)
				serverErrors <- err
			}
		}()
	} else {
		app.SystemLog.Info("[Server] SSH server is disabled in configuration")
	}

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或服务器错误
	select {
	case <-quit:
		app.SystemLog.Info("[Server] Received shutdown signal, shutting down server...")
	case err := <-serverErrors:
		app.SystemLog.Errorw("[Server] Server error occurred", "error", err)
		return err
	}

	// 停止接收信号通知
	signal.Stop(quit)
	close(quit)

	// 使用 context 通知服务器在配置的超时时间内完成当前请求
	shutdownTimeout := app.SystemConfig.Server.ShutdownTimeout
	if shutdownTimeout == 0 {
		// 如果配置文件中未设置，使用默认值 5 秒
		shutdownTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	app.SystemLog.Infow("[Server] Gracefully shutting down", "timeout", shutdownTimeout)

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		app.SystemLog.Errorw("[Server] HTTP server forced to shutdown", "error", err)
		// 不直接返回,继续清理其他资源
	} else {
		app.SystemLog.Info("[Server] HTTP server shutdown completed")
	}

	// 关闭 SSH 服务器(如果已启用)
	if app.SystemConfig.SSH.Enable {
		app.SystemLog.Info("[Server] Shutting down SSH server...")
		if err := app.SSHServer.Shutdown(ctx); err != nil {
			app.SystemLog.Errorw("[Server] SSH server forced to shutdown", "error", err)
		} else {
			app.SystemLog.Info("[Server] SSH server shutdown completed")
		}
	}

	// 停止所有定时任务
	app.SystemLog.Info("[Server] Stopping cron jobs...")
	cron.StopAll()
	return nil
}

// Register the server command.
func Register(app *kingpin.Application) {
	c := new(command)

	cmd := app.Command("server", "starts the server").
		Action(c.run)

	cmd.Arg("mode", "set the server mode").
		Default("dev").
		StringVar(&c.mode)
}
