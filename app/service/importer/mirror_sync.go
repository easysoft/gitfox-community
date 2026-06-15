package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	simpleJob "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/encrypt"
	providerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/provider"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

const (
	// JobTypeMirrorSync 镜像同步任务类型
	JobTypeMirrorSync = "mirror_sync"
	// MirrorSyncJobMaxDuration 镜像同步任务最大执行时间
	MirrorSyncJobMaxDuration = 10 * time.Minute
)

// MirrorSyncInput 镜像同步任务输入参数
type MirrorSyncInput struct {
	RepoID  int64  `json:"repoID"`
	JobType string `json:"jobType"` // 触发类型，定时任务或手动触发
}

// MirrorSyncHandler 镜像同步任务处理器，实现 pkgJobs.Handler
type MirrorSyncHandler struct {
	git           git.Interface
	svn           svn.Interface
	repoStore     repoStore.Interface
	providerStore providerStore.Interface
	encrypter     encrypt.Encrypter
}

// NewMirrorSyncHandler 创建镜像同步处理器
func NewMirrorSyncHandler(
	git git.Interface,
	svn svn.Interface,
	repoStore repoStore.Interface,
	providerStore providerStore.Interface,
	encrypter encrypt.Encrypter,
) *MirrorSyncHandler {
	return &MirrorSyncHandler{
		git:           git,
		svn:           svn,
		repoStore:     repoStore,
		providerStore: providerStore,
		encrypter:     encrypter,
	}
}

// CreateMirrorSyncJob 为镜像仓库创建单次同步任务。
// 创建任务前会将仓库状态更新为 syncing，任务执行成功后由 Handler 更新为 active。
func (r *Repository) CreateMirrorSyncJob(ctx context.Context, repoID int64) error {
	log := logger.GetLogger(ctx).With("repo.id", repoID)

	repo, err := r.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		log.Errorw("failed to find repo for mirror sync", "error", err)
		return errors.WithI18nKey(
			fmt.Errorf("failed to find repo for mirror sync: %w", err),
			i18n.RepoNotFound,
		)
	}

	if !repo.Mirror {
		log.Errorw("not allow to sync non-mirror repo")
		return errors.WithI18nKey(
			fmt.Errorf("not allow to sync non-mirror repo"),
			i18n.RepoMirrorSyncNotMirror,
		)
	}

	mirrorInput := MirrorSyncInput{
		RepoID:  repoID,
		JobType: JobTypeMirrorSync,
	}

	data, err := json.Marshal(mirrorInput)
	if err != nil {
		log.Errorw("failed to marshal mirror sync input", "error", err)
		return errors.WithI18nKey(
			fmt.Errorf("failed to marshal mirror sync input: %w", err),
			i18n.RepoMirrorSyncError,
		)
	}

	strData := strings.TrimSpace(string(data))
	encryptedData, err := r.encrypter.Encrypt(strData)
	if err != nil {
		log.Errorw("failed to encrypt mirror sync input", "error", err)
		return errors.WithI18nKey(
			fmt.Errorf("failed to encrypt mirror sync input: %w", err),
			i18n.RepoMirrorSyncError,
		)
	}

	jobName := fmt.Sprintf("mirror_sync_%d", repoID)
	jobData := base64.StdEncoding.EncodeToString(encryptedData)

	if err := r.scheduler.RunJob(ctx, jobName, JobTypeMirrorSync, jobData, int(MirrorSyncJobMaxDuration/time.Second)); err != nil {
		log.Errorw("failed to run mirror sync job", "error", err)
		return errors.WithI18nKey(
			fmt.Errorf("failed to run mirror sync job: %w", err),
			i18n.RepoMirrorSyncError,
		)
	}

	log.Infof("started mirror sync job %s", jobName)
	return nil
}

// GetMirrorSyncProgress 查询镜像同步任务的进度。
func (r *Repository) GetMirrorSyncProgress(ctx context.Context, repoID int64) (repoModel.Progress, error) {
	jobName := fmt.Sprintf("%s_%d", JobTypeMirrorSync, repoID)
	progress, err := r.scheduler.GetJobProgress(ctx, jobName)
	if err != nil {
		return repoModel.Progress{}, fmt.Errorf("failed to get mirror sync job progress: %w", err)
	}

	cronJobName := fmt.Sprintf("%s_%d", JobTypeMirrorSyncCron, repoID)
	cronProgress, err := r.scheduler.GetJobProgress(ctx, cronJobName)
	if err != nil {
		return repoModel.Progress{}, fmt.Errorf("failed to get mirror sync cron job progress: %w", err)
	}
	// 根据 LastExecuted 判断，返回最近一次执行的数据
	recentProgress := progress
	progressLastExec, _ := progress["lastExecuted"].(*datetime.Local)
	cronLastExec, _ := cronProgress["lastExecuted"].(*datetime.Local)

	// 如果 cron 任务最近执行过且比手动任务执行时间更晚，则使用 cron 的进度
	if cronLastExec != nil && (progressLastExec == nil || time.Time(*cronLastExec).After(time.Time(*progressLastExec))) {
		recentProgress = cronProgress
	}

	lastExecuted, _ := recentProgress["lastExecuted"].(*datetime.Local)
	return repoModel.Progress{
		Status:       recentProgress["state"].(simpleJob.State),
		Result:       recentProgress["result"].(string),
		Failure:      recentProgress["failure"].(string),
		LastExecuted: lastExecuted,
	}, nil
}

var _ pkgJobs.Handler = (*MirrorSyncHandler)(nil)

// Handle 执行镜像同步任务
func (h *MirrorSyncHandler) Handle(ctx context.Context, data string, reportFn pkgJobs.ProgressReporter) (result string, retErr error) {
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
	if repo.Status == constant.RepoStateSyncing {
		log.Info("repo is syncing")
		return "", nil
	}

	// 同步过程中出现任何错误，将仓库状态更新为 sync_failed
	defer func() {
		if retErr != nil {
			cleanupCtx := context.WithoutCancel(ctx)
			if _, updateErr := h.repoStore.UpdateOptLock(cleanupCtx, repo, func(repo *repoModel.Repo) error {
				repo.Status = constant.RepoStateSyncFailed
				return nil
			}); updateErr != nil {
				log.Errorw("failed to update repo status to sync_failed", "error", updateErr)
			}
		}
	}()

	if _, err := h.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
		repo.Status = constant.RepoStateSyncing
		return nil
	}); err != nil {
		return "", fmt.Errorf("failed to update repo status to syncing: %w", err)
	}

	provider, err := h.providerStore.GetByID(ctx, repo.ProviderID)
	if err != nil {
		return "", fmt.Errorf("failed to find provider for mirror sync: %w", err)
	}

	if provider.Type == string(constant.RepoProviderTypeSVN) {
		log.Info("mirror sync: start syncing SVN repository")

		svnConnectInfo := repoModel.SVNConnectInfo{}
		if err := json.Unmarshal(repo.Connector, &svnConnectInfo); err != nil {
			return "", fmt.Errorf("failed to unmarshal SVN connect info: %w", err)
		}

		passwordBytes, err := base64.StdEncoding.DecodeString(svnConnectInfo.Password)
		if err != nil {
			return "", fmt.Errorf("failed to decode SVN password: %w", err)
		}
		svnConnectInfo.Password = string(passwordBytes)

		if err := h.syncSvnRepo(ctx, repo.GitUID, svnConnectInfo.User, svnConnectInfo.Password, log); err != nil {
			return "", err
		}
	} else {
		log.Info("mirror sync: start syncing repository")

		gitConnectInfo := repoModel.GitConnectInfo{}
		if err := json.Unmarshal(repo.Connector, &gitConnectInfo); err != nil {
			return "", fmt.Errorf("failed to unmarshal Git connect info: %w", err)
		}
		cloneURL := fmt.Sprintf("%s/%s", provider.URL, gitConnectInfo.Slug)
		if err := h.syncGitRepo(ctx, repo.GitUID, cloneURL, log); err != nil {
			return "", err
		}
	}

	if _, err := h.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
		repo.Status = constant.RepoStateActive
		return nil
	}); err != nil {
		return "", fmt.Errorf("failed to update repo status after mirror sync: %w", err)
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

// syncGitRepo 执行 Git 仓库的镜像同步。
func (h *MirrorSyncHandler) syncGitRepo(ctx context.Context, gitUID, cloneURL string, log *logger.SugaredMateLogger) error {
	synced, err := h.git.MirrorSyncRepository(ctx, &git.MirrorSyncParams{
		ReadParams: git.ReadParams{
			RepoUID: gitUID,
		},
		CloneURL: cloneURL,
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

func (h *MirrorSyncHandler) syncSvnRepo(ctx context.Context, svnUID, username, password string, log *logger.SugaredMateLogger) error {
	path := h.svn.ResolveRepositoryPaths(svnUID)

	err := h.svn.Sync(ctx, &svn.SyncParams{
		TargetRepoPath: path.RepoPath,
		SourceUsername: username,
		SourcePassword: password,
	})
	if err != nil {
		return fmt.Errorf("failed to sync SVN repository: %w", err)
	}
	err = h.svn.SyncCopyRevProps(ctx, &svn.SyncParams{
		TargetRepoPath: path.RepoPath,
		SourceUsername: username,
		SourcePassword: password,
	})
	if err != nil {
		return fmt.Errorf("failed to sync SVN repository revision properties: %w", err)
	}
	err = h.svn.Update(ctx, &svn.UpdateParams{
		WorkingCopyPath: path.WorkingCopyPath,
	})
	if err != nil {
		return fmt.Errorf("failed to update SVN working copy: %w", err)
	}

	log.Info("mirror sync: successfully synced")
	return nil
}
