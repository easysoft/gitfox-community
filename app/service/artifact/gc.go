package artifact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

// TriggerGarbageCollectContainer 容器镜像gc回收
func (s *Service) TriggerGarbageCollectContainer(ctx context.Context) (*model.JobUIDResponse, error) {
	return s.triggerGarbageCollect(ctx, JobTypeArtifactContainerGC)
}

// TriggerGarbageCollectSoftRemove  其他资源gc回收 / raw、generic、helm
func (s *Service) TriggerGarbageCollectSoftRemove(ctx context.Context) (*model.JobUIDResponse, error) {
	return s.triggerGarbageCollect(ctx, JobTypeArtifactSoftRemoveGC)
}

func (s *Service) triggerGarbageCollect(ctx context.Context, jobType string) (*model.JobUIDResponse, error) {
	if s.scheduler == nil {
		return nil, pkgErrors.WithI18nKey(
			errors.New("artifact job scheduler is not configured"),
			i18n.ArtifactGcSchedulerNotConfigured,
		)
	}
	jobName := fmt.Sprintf("%s:%d", jobType, time.Now().UnixNano())
	payload, err := json.Marshal(map[string]any{"retention_time": 0})
	if err != nil {
		return nil, err
	}

	if err = s.scheduler.RunJob(ctx, jobName, jobType, string(payload), ArtifactGCJobMaxDurationSeconds); err != nil {
		return nil, pkgErrors.WithI18nKey(err, i18n.ArtifactGcTriggerError)
	}
	return &model.JobUIDResponse{UID: jobName}, nil
}
