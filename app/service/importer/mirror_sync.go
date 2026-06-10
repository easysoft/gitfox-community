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
	// MirrorSyncJobMaxDuration 镜像同步任务最大执行时间
	MirrorSyncJobMaxDuration = 10 * time.Minute
)

// MirrorSyncInput 镜像同步任务输入参数（加密存储在 job.Data 中）
type MirrorSyncInput struct {
	RepoID int64 `json:"repoID"`

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

	if err := h.syncGitRepo(ctx, repo, input, log); err != nil {
		return "", err
	}

	return "", nil
}

// syncGitRepo 执行 Git 仓库的镜像同步。
func (h *MirrorSyncHandler) syncGitRepo(ctx context.Context, repo *repoModel.Repo, input MirrorSyncInput, log *logger.SugaredMateLogger) error {
	synced, err := h.git.MirrorSyncRepository(ctx, &git.MirrorSyncParams{
		ReadParams: git.ReadParams{
			RepoUID: repo.GitUID,
		},
		CloneURL: input.CloneURL,
		Prune:    true,
	})
	if err != nil {
		return fmt.Errorf("mirror sync failed: %w", err)
	}

	if synced {
		log.Info("mirror sync: successfully synced")
	} else {
		log.Info("mirror sync: no updates")
	}

	return nil
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

// CreateMirrorSyncJob 为镜像仓库创建单次同步任务。
func (r *Repository) CreateMirrorSyncJob(ctx context.Context, repoID int64, cloneURL, user, auth string, log *logger.SugaredMateLogger,
) error {
	mirrorInput := MirrorSyncInput{
		RepoID:   repoID,
		User:     user,
		Auth:     auth,
		CloneURL: cloneURL,
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

	jobName := fmt.Sprintf("mirror_sync_%d", repoID)
	jobData := base64.StdEncoding.EncodeToString(encryptedData)

	if err := r.scheduler.RunJob(ctx, jobName, JobTypeMirrorSync, jobData, int(MirrorSyncJobMaxDuration/time.Second)); err != nil {
		log.Errorw("failed to run mirror sync job", "error", err)
		return err
	}

	log.Infof("started mirror sync job %s", jobName)
	return nil
}
