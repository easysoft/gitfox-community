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
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/logger"

	// artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact" // TODO: 制品库 GC 定时任务禁用时暂注释
	cronSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/cron"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
)

type command struct {
	mode     string
	enableCI bool
}

func (c *command) run(*kingpin.ParseContext) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统中断信号（拦截 Ctrl+C / Systemd 停止）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		// 收到信号后通过 context 触发内部停止
		cancel()
	}()

	return RunServerWithContext(ctx, c.mode, c.enableCI, nil)
}

// RunServerWithContext 是 GitFox 的通用启动引擎。
// 它通过 ctx 控制生命周期：当 ctx 被 cancel 时，执行优雅关机。
// onReady 是一个可选的回调函数，当网络端口挂载完毕、正式开始监听时触发。
func RunServerWithContext(ctx context.Context, mode string, enableCI bool, onReady func()) error {
	app, cleanup, err := NewApp(mode)
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

	// 1. 数据迁移
	app.runMigrations()

	// 2. 注册后台任务处理器
	app.registerJobHandlers()

	// 3. 注册定时任务
	app.registerRecurringJobs(ctx)

	// 启动 Job Scheduler
	app.SystemLog.Info("[Server] Starting job scheduler...")
	jobSchedulerCtx, cancelJobScheduler := context.WithCancel(context.Background())
	defer cancelJobScheduler() // Ensure context is canceled on all return paths
	jobSchedulerErrors := make(chan error, 1)

	go func() {
		if err := app.JobScheduler.Run(jobSchedulerCtx); err != nil && err != context.Canceled {
			app.SystemLog.Errorw("[Server] Job scheduler error", "error", err)
			jobSchedulerErrors <- err
		}
	}()
	app.SystemLog.Info("[Server] Job scheduler started")

	// 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", app.SystemConfig.Server.Host, app.SystemConfig.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router.NewRouter(mode, app.APILog.SugaredMateLogger, app.Controller, app.ArtifactAuthorizer, app.ContainerHandler, app.SystemConfig),
	}

	// 创建错误通道用于捕获服务器启动错误
	serverErrors := make(chan error, 3) // 增加容量以支持 HTTP、SSH 和 JobScheduler

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

	pollerCtx, cancelPoller := context.WithCancel(context.Background())
	defer cancelPoller()
	if enableCI {
		// start poller for CI build executions (non-critical for API)
		go app.Poller.Poll(
			logger.WithWrappedZerolog(pollerCtx),
			int(app.SystemConfig.Ci.ParallelWorkers),
		)
	}

	// 当主业务模块挂载成功时，通知外部组件
	if onReady != nil {
		onReady()
	}

	// 阻塞等待：等待外部取消请求或内部子服务抛出致命错误
	select {
	case <-ctx.Done():
		app.SystemLog.Info("[Server] Context canceled, initiating graceful shutdown...")
	case err := <-serverErrors:
		app.SystemLog.Errorw("[Server] Server error occurred", "error", err)
		return err
	case err := <-jobSchedulerErrors:
		app.SystemLog.Errorw("[Server] Job scheduler error occurred", "error", err)
		return err
	}

	// 使用 context 通知服务器在配置的超时时间内完成当前请求
	shutdownTimeout := app.SystemConfig.Server.ShutdownTimeout
	if shutdownTimeout == 0 {
		// 如果配置文件中未设置，使用默认值 5 秒
		shutdownTimeout = 5 * time.Second
	}
	// 使用新的带超时的 context 来执行优雅关机操作
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	app.SystemLog.Infow("[Server] Gracefully shutting down", "timeout", shutdownTimeout)

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		app.SystemLog.Errorw("[Server] HTTP server forced to shutdown", "error", err)
		// 不直接返回,继续清理其他资源
	} else {
		app.SystemLog.Info("[Server] HTTP server shutdown completed")
	}

	// 关闭 SSH 服务器(如果已启用)
	if app.SystemConfig.SSH.Enable {
		app.SystemLog.Info("[Server] Shutting down SSH server...")
		if err := app.SSHServer.Shutdown(shutdownCtx); err != nil {
			app.SystemLog.Errorw("[Server] SSH server forced to shutdown", "error", err)
		} else {
			app.SystemLog.Info("[Server] SSH server shutdown completed")
		}
	}

	// 停止 CI Poller
	app.SystemLog.Info("[Server] Stopping CI poller...")
	cancelPoller()
	app.SystemLog.Info("[Server] CI poller shutdown completed")

	// 关闭事件系统（取消所有事件消费者）
	app.SystemLog.Info("[Server] Stopping event system...")
	app.EventContext.Cancel()
	app.SystemLog.Info("[Server] Event system shutdown completed")

	// 停止 Job Scheduler
	app.SystemLog.Info("[Server] Stopping job scheduler...")
	app.JobScheduler.WaitJobsDone(shutdownCtx)
	app.SystemLog.Info("[Server] Job scheduler shutdown completed")

	return nil
}

// runMigrations 执行启动时的数据迁移任务。
func (app *App) runMigrations() {
	if err := app.ScanRuleSvc.Migrate(context.Background()); err != nil {
		app.SystemLog.Errorw("[Server] Failed to migrate scan rules", "error", err)
		panic(err)
	}

	if err := app.PipelineService.Migrate(context.Background(), app.SystemConfig.Ci.PluginsZipUrl); err != nil {
		app.SystemLog.Errorw("[Server] Failed to migrate plugins", "error", err)
		panic(err)
	}

	if err := app.TokenStore.SetDevopsToken(context.Background()); err != nil {
		app.SystemLog.Errorw("[Server] Failed to set devops token", "error", err)
		panic(err)
	}
}

// registerJobHandlers 注册后台任务处理器（Pipeline Cron、Scan Cron、Artifact GC）。
func (app *App) registerJobHandlers() {
	if err := cronSvc.RegisterPipelineHandler(
		app.JobScheduler, app.Triggerer, app.PipelineService); err != nil {
		app.SystemLog.Errorw("[Server] Failed to register pipeline cron handler", "error", err)
		panic(err)
	}

	if err := cronSvc.RegisterScanHandler(
		app.JobScheduler, app.ScanTaskSvc, app.ScanPlanSvc, app.ScanTriggerStore, app.RepoService); err != nil {
		app.SystemLog.Errorw("[Server] Failed to register scan cron handler", "error", err)
		panic(err)
	}

	if err := cronSvc.RegisterArtifactGCHandler(app.JobScheduler, app.ArtifactService); err != nil {
		app.SystemLog.Errorw("[Server] Failed to register artifact GC handler", "error", err)
		panic(err)
	}
}

// registerRecurringJobs 注册定时调度任务。
// 制品库 GC 的 cron 间隔由配置 artifact.gcInterval 驱动（默认 720h = 每30天），
// 三个 GC 任务通过 minuteOffset 错峰执行，避免同时竞争资源。
func (app *App) registerRecurringJobs(ctx context.Context) {
	// TODO: 制品库 GC 定时任务暂时禁用，待回收站策略确认后恢复
	// app.SystemConfig.Artifact.Sanitize()
	// artCfg := &app.SystemConfig.Artifact
	// app.SystemLog.Infow("[Server] Artifact GC config",
	// 	"recycleBinRetention", artCfg.RecycleBinRetention,
	// 	"gcInterval", artCfg.GCInterval,
	// )

	// gcJobs := []struct {
	// 	name    string
	// 	jobName string
	// 	jobType string
	// 	offset  int // minuteOffset，用于错峰
	// }{
	// 	{"container GC", artifactSvc.JobNameArtifactContainerGC, artifactSvc.JobTypeArtifactContainerGC, 0},
	// 	{"soft-remove GC", artifactSvc.JobNameArtifactSoftRemoveGC, artifactSvc.JobTypeArtifactSoftRemoveGC, 15},
	// 	{"expired view GC", artifactSvc.JobNameArtifactExpiredViewGC, artifactSvc.JobTypeArtifactExpiredViewGC, 30},
	// }

	// for _, j := range gcJobs {
	// 	cronExpr := artCfg.GCCronExpr(3, j.offset)
	// 	if err := app.JobScheduler.AddRecurringWithData(
	// 		ctx,
	// 		j.jobName,
	// 		j.jobType,
	// 		"{}",
	// 		cronExpr,
	// 		time.Duration(artifactSvc.ArtifactGCJobMaxDurationSeconds)*time.Second,
	// 	); err != nil {
	// 		app.SystemLog.Errorw("[Server] Failed to register artifact "+j.name+" recurring job", "error", err)
	// 		panic(err)
	// 	}
	// 	app.SystemLog.Infow("[Server] Artifact "+j.name+" recurring job registered", "cron", cronExpr)
	// }

	// TODO: 重新协商方案 启动时对账：确保所有 trigger 的 cron job 与 DB 一致（修复上次异常退出可能导致的不一致）
	// app.SystemLog.Info("[Server] Reconciling pipeline cron jobs...")
	// if err := app.PipelineService.ReconcileCronJobs(context.Background()); err != nil {
	// 	app.SystemLog.Errorw("[Server] Failed to reconcile pipeline cron jobs", "error", err)
	// 	panic(err)
	// }
	// app.SystemLog.Info("[Server] Pipeline cron jobs reconciled")
}

// Register the server command.
func Register(app *kingpin.Application) {
	c := new(command)

	cmd := app.Command(clii18n.T(i18n.CliServer), clii18n.T(i18n.CliServerDescription)).
		Default().
		Action(c.run).Hidden()

	cmd.Arg("mode", clii18n.T(i18n.CliServerModeArg)).
		Default("dev").
		StringVar(&c.mode)

	cmd.Flag("enable-ci", clii18n.T(i18n.CliServerEnableCiFlag)).
		Default("true").
		Envar("ENABLE_CI").
		BoolVar(&c.enableCI)
}
