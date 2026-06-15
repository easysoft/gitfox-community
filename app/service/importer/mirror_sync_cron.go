package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

const (
	// JobTypeMirrorSync 镜像同步任务类型
	JobTypeMirrorSyncCron = "mirror_sync_cron"
	// DefaultMirrorSyncInterval 默认镜像同步间隔（分钟）
	DefaultMirrorSyncInterval = 10
	// MirrorSyncCronJobMaxDuration 镜像同步任务最大执行时间
	MirrorSyncCronJobMaxDuration = 10 * time.Minute
)

// SyncIntervalToCron 将同步间隔（分钟）转换为 cron 表达式。
func SyncIntervalToCron(minutes uint) string {
	if minutes < 1 {
		minutes = DefaultMirrorSyncInterval
	}
	if minutes < 60 {
		return fmt.Sprintf("*/%d * * * *", minutes)
	}
	hours := minutes / 60
	remainder := minutes % 60
	if remainder == 0 {
		if hours == 1 {
			return "0 * * * *"
		}
		return fmt.Sprintf("0 */%d * * *", hours)
	}
	return fmt.Sprintf("*/%d * * * *", minutes)
}

// CreateMirrorSyncCronJob 为镜像仓库创建定时同步任务。
func (r *Repository) CreateMirrorSyncCronJob(ctx context.Context, repoID int64, log *logger.SugaredMateLogger) error {
	mirrorInput := MirrorSyncInput{
		RepoID:  repoID,
		JobType: JobTypeMirrorSyncCron,
	}

	data, err := json.Marshal(mirrorInput)
	if err != nil {
		log.Errorw("failed to marshal mirror sync input", "error", err)
		return err
	}

	strData := strings.TrimSpace(string(data))
	encryptedData, err := r.encrypter.Encrypt(strData)
	if err != nil {
		log.Errorw("failed to encrypt mirror sync input", "error", err)
		return err
	}

	jobName := fmt.Sprintf("mirror_sync_cron_%d", repoID)
	cronExpr := SyncIntervalToCron(DefaultMirrorSyncInterval)
	jobData := base64.StdEncoding.EncodeToString(encryptedData)

	if err := r.scheduler.RunRecurringJob(ctx, jobName, JobTypeMirrorSyncCron, cronExpr, jobData, MirrorSyncCronJobMaxDuration); err != nil {
		log.Errorw("failed to create mirror sync recurring job", "error", err)
		return err
	}

	log.Infof("created mirror sync recurring job %s with cron %s", jobName, cronExpr)
	return nil
}
