package server

import (
	"context"
	"fmt"
	"log"
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
	app, cleanup, err := NewAppServer(c.mode)
	if err != nil {
		log.Fatalf("[Server] Failed to initialize app server: %v", err)
	}
	defer func() {
		log.Println("[Server] Cleaning up resources...")
		cleanup()
	}()

	// 启动定时任务管理器
	cronManager := cron.NewCronManager()

	// 示例：添加系统健康检查定时任务：每一分钟打印一条系统日志
	_, err = cronManager.AddJob("system-health-check", "0 * * * * *", func() {
		(*app.SystemLog).Infow("System health check", "status", "running")
	})
	if err != nil {
		log.Printf("[Server] Failed to add cron job: %v", err)
	} else {
		log.Println("[Server] System health check cron job started (runs every minute)")
	}

	// 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", app.SystemConfig.Server.Host, app.SystemConfig.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router.NewRouter(c.mode, app.APILog, app.Handlers),
	}

	// 创建错误通道用于捕获服务器启动错误
	serverErrors := make(chan error, 1)

	// 在 goroutine 中启动服务器
	go func() {
		log.Printf("[Server] Starting HTTP server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[Server] HTTP server error: %v", err)
			serverErrors <- err
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或服务器错误
	select {
	case <-quit:
		log.Println("[Server] Received shutdown signal, shutting down server...")
	case err := <-serverErrors:
		log.Printf("[Server] Server error occurred: %v", err)
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

	log.Printf("[Server] Gracefully shutting down (timeout: %v)...", shutdownTimeout)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[Server] Server forced to shutdown: %v", err)
		// 不直接返回,继续清理其他资源
	} else {
		log.Println("[Server] HTTP server shutdown completed")
	}

	// 停止所有定时任务
	log.Println("[Server] Stopping cron jobs...")
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
