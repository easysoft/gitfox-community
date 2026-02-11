package jobs

import (
	"context"
	"fmt"
	"time"

	runnerModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/runner"
	runnerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/runner"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"

	"github.com/rs/zerolog/log"
)

const (
	jobUIDRunnerHeartbeat         = "gitfox:runner:heartbeat-check"
	jobTypeRunnerHeartbeat        = "gitfox:runner:heartbeat-check"
	jobCronRunnerHeartbeat        = "* * * * *"      // 每分钟执行一次
	jobMaxDurationRunnerHeartbeat = 30 * time.Second // 最大执行时间
	heartbeatTimeoutSeconds       = 30               // 心跳超时阈值：30 秒
)

type runnerHeartbeatJob struct {
	runnerStore runnerStore.Interface
}

func newRunnerHeartbeatJob(runnerStore runnerStore.Interface, mxManager lock.MutexManager, scheduler *Scheduler) *runnerHeartbeatJob {
	return &runnerHeartbeatJob{
		runnerStore: runnerStore,
	}
}

// Handle 检测 runner 心跳并更新离线状态
func (j *runnerHeartbeatJob) Handle(ctx context.Context, _ string, reporter pkgJobs.ProgressReporter) (string, error) {
	log.Ctx(ctx).Trace().Msg("starting runner heartbeat check")

	// 设置初始进度
	if _, err := reporter(10, "initializing"); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to report progress")
	}

	// 创建过滤器，查询所有未删除且在线的 runner
	filter := &runnerModel.ReqFilter{
		Page:     1,
		PageSize: 10000,
		Online:   "online", // 只查询在线状态的 runner，避免重复处理已离线的
	}

	// 获取所有在线状态的 runner
	res, err := j.runnerStore.List(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("failed to list runners: %w", err)
	}

	log.Ctx(ctx).Debug().Int("total_runners", len(res.Data)).Msg("found runners for heartbeat check")

	// 设置进度
	if _, err := reporter(30, "scanning runners"); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to report progress")
	}

	// 当前时间戳
	now := time.Now()
	offlineCount := 0

	// 检查每个 runner 的心跳
	for _, runner := range res.Data {
		// 由于已经过滤了在线状态，无需再次检查 offline 状态

		// 检查心跳时间
		if runner.HeartBeat == nil {
			// 如果没有心跳时间，标记为离线
			if err := j.updateRunnerOffline(ctx, int64(runner.ID)); err != nil {
				log.Ctx(ctx).Error().
					Err(err).
					Uint("runner_id", runner.ID).
					Str("runner_name", runner.Name).
					Msg("failed to update runner to offline (no heartbeat)")
				continue
			}
			offlineCount++
			log.Ctx(ctx).Debug().
				Uint("runner_id", runner.ID).
				Str("runner_name", runner.Name).
				Msg("runner marked offline: no heartbeat")
			continue
		}

		// 计算心跳时间差
		heartbeatTime := time.Unix(int64(*runner.HeartBeat), 0)
		timeSinceHeartbeat := now.Sub(heartbeatTime)

		// 如果超过30秒没有心跳，标记为离线
		if timeSinceHeartbeat > heartbeatTimeoutSeconds*time.Second {
			if err := j.updateRunnerOffline(ctx, int64(runner.ID)); err != nil {
				log.Ctx(ctx).Error().
					Err(err).
					Uint("runner_id", runner.ID).
					Str("runner_name", runner.Name).
					Str("time_since_heartbeat", timeSinceHeartbeat.String()).
					Msg("failed to update runner to offline")
				continue
			}
			offlineCount++
			log.Ctx(ctx).Debug().
				Uint("runner_id", runner.ID).
				Str("runner_name", runner.Name).
				Str("time_since_heartbeat", timeSinceHeartbeat.String()).
				Msg("runner marked offline: heartbeat timeout")
		}
	}

	// 设置最终进度
	if _, err := reporter(100, "completed"); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to report progress")
	}

	result := fmt.Sprintf("checked %d runners, marked %d as offline", len(res.Data), offlineCount)
	log.Ctx(ctx).Info().
		Int("total_checked", len(res.Data)).
		Int("marked_offline", offlineCount).
		Msg("completed runner heartbeat check")

	return result, nil
}

// updateRunnerOffline 将指定的 runner 标记为离线
func (j *runnerHeartbeatJob) updateRunnerOffline(ctx context.Context, runnerID int64) error {
	runnerInfo := &runnerModel.Runner{
		ID:     uint(runnerID),
		Online: "offline",
	}

	return j.runnerStore.Update(ctx, runnerInfo)
}
