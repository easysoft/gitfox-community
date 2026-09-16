package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cronUtil "github.com/easysoft/gitfox-community/pkg/utils/cron"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const (
	// JobTypeMirrorSync 镜像同步任务类型
	JobTypeMirrorSyncCron = "mirror_sync_cron"
	// MirrorSyncCronJobMaxDuration 镜像同步任务最大执行时间
	MirrorSyncCronJobMaxDuration = 10 * time.Minute
)

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
	cronExpr := cronUtil.SyncIntervalToCron(cronUtil.DefaultMirrorSyncInterval)
	jobData := base64.StdEncoding.EncodeToString(encryptedData)

	if err := r.scheduler.RunRecurringJob(ctx, jobName, JobTypeMirrorSyncCron, cronExpr, jobData, MirrorSyncCronJobMaxDuration); err != nil {
		log.Errorw("failed to create mirror sync recurring job", "error", err)
		return err
	}

	log.Infof("created mirror sync recurring job %s with cron %s", jobName, cronExpr)
	return nil
}
