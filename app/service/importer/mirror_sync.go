// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
	simpleJob "github.com/easysoft/gitfox-community/app/model/jobs"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/app/service/encrypt"
	providerStore "github.com/easysoft/gitfox-community/app/store/provider"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	pkgJobs "github.com/easysoft/gitfox-community/pkg/jobs"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/svn"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
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
// 手动任务和 cron 任务只要有一个存在即可查询；两个都存在时，返回最近一次执行的数据。
func (r *Repository) GetMirrorSyncProgress(ctx context.Context, repoID int64) (repoModel.Progress, error) {
	jobName := fmt.Sprintf("%s_%d", JobTypeMirrorSync, repoID)
	cronJobName := fmt.Sprintf("%s_%d", JobTypeMirrorSyncCron, repoID)

	progress, progressErr := r.scheduler.GetJobProgress(ctx, jobName)
	cronProgress, cronErr := r.scheduler.GetJobProgress(ctx, cronJobName)

	// 两种 job 都查询失败时才返回 error
	if progressErr != nil && cronErr != nil {
		return repoModel.Progress{}, fmt.Errorf("failed to get mirror sync progress: progress=%w, cron=%w", progressErr, cronErr)
	}

	// 根据 LastExecuted 选择最近一次执行的进度数据
	recentProgress := pickRecentProgress(progress, progressErr, cronProgress, cronErr)

	lastExecuted, _ := recentProgress["lastExecuted"].(*datetime.Local)
	nextExecuted, _ := recentProgress["nextExecuted"].(*datetime.Local)
	status, _ := recentProgress["state"].(simpleJob.State)
	result, _ := recentProgress["result"].(string)
	failure, _ := recentProgress["failure"].(string)
	return repoModel.Progress{
		Status:       status,
		Result:       result,
		Failure:      failure,
		LastExecuted: lastExecuted,
		NextExecuted: nextExecuted,
	}, nil
}

// pickRecentProgress 根据两个 job 的查询结果和 lastExecuted 时间，返回最近执行的进度。
// 只有一个存在时直接返回该 job；两个都存在时比较 lastExecuted 取较新的。
func pickRecentProgress(progress map[string]interface{}, progressErr error, cronProgress map[string]interface{}, cronErr error) map[string]interface{} {
	if progressErr != nil {
		return cronProgress
	}
	if cronErr != nil {
		return progress
	}

	progressLastExec, _ := progress["lastExecuted"].(*datetime.Local)
	cronLastExec, _ := cronProgress["lastExecuted"].(*datetime.Local)
	if cronLastExec != nil && (progressLastExec == nil || time.Time(*cronLastExec).After(time.Time(*progressLastExec))) {
		return cronProgress
	}
	return progress
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
	if strings.HasPrefix(repo.GitUID, constant.MigrateRepoGitUID) {
		log.Infow("mirror sync: skip repo with migrate placeholder gitUID", "repo.gitUID", repo.GitUID)
		return "", nil
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
	err := h.svn.Sync(ctx, &svn.SyncParams{
		SvnUID:         svnUID,
		SourceUsername: username,
		SourcePassword: password,
	})
	if err != nil {
		return fmt.Errorf("failed to sync SVN repository: %w", err)
	}
	err = h.svn.SyncCopyRevProps(ctx, &svn.SyncParams{
		SvnUID:         svnUID,
		SourceUsername: username,
		SourcePassword: password,
	})
	if err != nil {
		return fmt.Errorf("failed to sync SVN repository revision properties: %w", err)
	}
	err = h.svn.Update(ctx, &svn.UpdateParams{
		SvnUID: svnUID,
	})
	if err != nil {
		return fmt.Errorf("failed to update SVN working copy: %w", err)
	}

	log.Info("mirror sync: successfully synced")
	return nil
}
