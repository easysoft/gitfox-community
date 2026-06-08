package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/encrypt"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

const (
	// JobTypeMirrorSync 镜像同步任务类型
	JobTypeMirrorSync = "mirror_sync"
	// DefaultMirrorSyncInterval 默认镜像同步间隔（分钟）
	DefaultMirrorSyncInterval = 60
	// MirrorSyncJobMaxDuration 镜像同步任务最大执行时间
	MirrorSyncJobMaxDuration = 60 * time.Minute
)

// MirrorSyncInput 镜像同步任务输入参数（加密存储在 job.Data 中）
type MirrorSyncInput struct {
	RepoID   int64  `json:"repoID"`
	User     string `json:"user"`
	Auth     string `json:"auth"`
	CloneURL string `json:"cloneURL"`
}

// repoStoreMirrorSync 定义 mirror sync 所需的 repo store 方法
type repoStoreMirrorSync interface {
	GetByID(ctx context.Context, id uint) (*repoModel.Repo, error)
}

// MirrorSyncHandler 镜像同步任务处理器，实现 pkgJobs.Handler
type MirrorSyncHandler struct {
	git       git.Interface
	repoStore repoStoreMirrorSync
	encrypter encrypt.Encrypter
}

// NewMirrorSyncHandler 创建镜像同步处理器
func NewMirrorSyncHandler(
	g git.Interface,
	repoStore repoStoreMirrorSync,
	encrypter encrypt.Encrypter,
) *MirrorSyncHandler {
	return &MirrorSyncHandler{
		git:       g,
		repoStore: repoStore,
		encrypter: encrypter,
	}
}

var _ pkgJobs.Handler = (*MirrorSyncHandler)(nil)

// Handle 执行镜像同步任务
func (h *MirrorSyncHandler) Handle(ctx context.Context, data string, reportFn pkgJobs.ProgressReporter) (string, error) {
	input, err := h.getJobInput(data)
	if err != nil {
		return "", err
	}

	log := logger.GetLogger(ctx).With(
		"job.Type", JobTypeMirrorSync,
		"repo.id", input.RepoID,
	)

	repo, err := h.repoStore.GetByID(ctx, uint(input.RepoID))
	if err != nil {
		return "", fmt.Errorf("failed to find repo for mirror sync: %w", err)
	}

	log.Info("mirror sync: start syncing repository")

	synced, err := h.git.MirrorSyncRepository(ctx, &git.MirrorSyncParams{
		ReadParams: git.ReadParams{
			RepoUID: repo.GitUID,
		},
		CloneURL: input.CloneURL,
		Prune:    true,
	})
	if err != nil {
		return "", fmt.Errorf("mirror sync failed: %w", err)
	}

	if synced {
		log.Info("mirror sync: successfully synced")
	} else {
		log.Info("mirror sync: no updates")
	}

	return "", nil
}

// getJobInput 解密并解析任务输入
func (h *MirrorSyncHandler) getJobInput(data string) (MirrorSyncInput, error) {
	encrypted, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return MirrorSyncInput{}, fmt.Errorf("failed to base64 decode mirror sync job input: %w", err)
	}

	decrypted, err := h.encrypter.Decrypt(encrypted)
	if err != nil {
		return MirrorSyncInput{}, fmt.Errorf("failed to decrypt mirror sync job input: %w", err)
	}

	var input MirrorSyncInput
	err = json.NewDecoder(strings.NewReader(decrypted)).Decode(&input)
	if err != nil {
		return MirrorSyncInput{}, fmt.Errorf("failed to unmarshal mirror sync job input json: %w", err)
	}

	return input, nil
}

// MirrorSyncJobName 生成镜像同步任务的唯一名称
func MirrorSyncJobName(repoID int64) string {
	return fmt.Sprintf("mirror_sync_%d", repoID)
}

// SyncIntervalToCron 将同步间隔（分钟）转换为 cron 表达式
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
