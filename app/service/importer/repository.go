package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	eventModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/event"
	simpleJob "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/encrypt"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	notifyStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/notify"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	urlUtil "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

const (
	importJobMaxRetries  = 0
	importJobMaxDuration = 45 * time.Minute
)

var (
	// ErrNotFound is returned if no import data was found.
	ErrNotFound = errors.NotFound.New("import not found")
)

type Repository struct {
	defaultBranch string
	urlProvider   urlUtil.Provider
	git           git.Interface
	repoStore     repoStore.Interface
	notifyStore   notifyStore.Interface
	encrypter     encrypt.Encrypter
	scheduler     *jobs.Scheduler
	sseStreamer   sse.Streamer
}

var _ Interface = (*Repository)(nil)

// NewRepoImporterSvc 创建 RepoService 实例
func NewRepoImporterSvc(
	defaultBranch string,
	urlProvider urlUtil.Provider,
	git git.Interface,
	repoStore repoStore.Interface,
	notifyStore notifyStore.Interface,
	encrypter encrypt.Encrypter,
	scheduler *jobs.Scheduler,
	sseStreamer sse.Streamer,
) *Repository {
	return &Repository{
		defaultBranch: defaultBranch,
		urlProvider:   urlProvider,
		git:           git,
		repoStore:     repoStore,
		notifyStore:   notifyStore,
		encrypter:     encrypter,
		scheduler:     scheduler,
		sseStreamer:   sseStreamer,
	}
}

func (r *Repository) Create(ctx context.Context, session *auth.Session, req *repoModel.Repo) (*repoModel.Repo, error) {
	repo := &repoStore.Repository{
		SpaceID:       req.SpaceID,
		Identifier:    req.Identifier,
		GitUID:        req.GitUID,
		Desc:          req.Desc,
		Product:       req.Product,
		CreatedBy:     session.User.Account,
		DefaultBranch: req.DefaultBranch,
		Status:        req.Status,
		ACL:           req.ACL,
		Mirror:        req.Mirror,
	}

	repoObj, err := r.repoStore.Create(ctx, repo)
	if err != nil {
		return nil, err
	}
	repoObj.GitSSHURL = r.urlProvider.GenerateGITCloneSSHURL(ctx, repoObj.Path)
	repoObj.GitURL = r.urlProvider.GenerateGITCloneURL(ctx, repoObj.Path)
	return repoObj, nil
}

func (r *Repository) LoadRepositoryFromProvider(
	ctx context.Context,
	provider repoModel.Provider,
	slug string,
) (RepositoryInfo, repoModel.Provider, error) {
	if slug == "" {
		return RepositoryInfo{}, provider, errors.WithI18nKey(
			errors.BadRequest.New("provider repository identifier is missing"),
			i18n.RepoProviderRepoIdentifierMissing,
		)
	}

	scmClient, err := getScmClientWithTransport(provider, slug, false)
	if err != nil {
		return RepositoryInfo{}, provider, errors.WithI18nKey(
			errors.BadRequest.Wrap(err, "could not create client"),
			i18n.RepoProviderScmClientCreateError,
		)
	}

	scmRepo, scmResp, err := scmClient.Repositories.Find(ctx, slug)
	if err = convertImportError(provider, slug, scmResp, err); err != nil {
		return RepositoryInfo{}, provider, errors.WithI18nKey(err, i18n.RepoImportError)
	}

	return RepositoryInfo{
		Space:         scmRepo.Namespace,
		Identifier:    scmRepo.Name,
		CloneURL:      scmRepo.Clone,
		DefaultBranch: scmRepo.Branch,
		IsPublic:      !scmRepo.Private,
	}, provider, nil
}

func (r *Repository) Run(ctx context.Context, provider repoModel.Provider, repo *repoModel.Repo, cloneURL string, mirror bool, noticeLang string) error {
	jobDef, err := r.getJobDef(JobIDFromRepoID(int64(repo.ID)), Input{
		RepoID:     int64(repo.ID),
		GitUser:    provider.Username,
		GitPass:    provider.Password,
		CloneURL:   cloneURL,
		Mirror:     mirror,
		NoticeLang: noticeLang,
	})
	if err != nil {
		return err
	}

	return r.scheduler.RunJob(ctx, jobDef.Name, jobDef.Type, jobDef.Data, int(jobDef.Timeout/time.Second))
}

func (r *Repository) GetProgress(ctx context.Context, repoID uint) (repoModel.Progress, error) {
	progress, err := r.scheduler.GetJobProgress(ctx, JobIDFromRepoID(int64(repoID)))
	if err != nil {
		return repoModel.Progress{}, fmt.Errorf("failed to get job progress: %w", err)
	}

	return repoModel.Progress{
		Status:  progress["state"].(simpleJob.State),
		Result:  progress["result"].(string),
		Failure: progress["failure"].(string),
	}, nil
}

func (r *Repository) getJobDef(jobUID string, input Input) (simpleJob.Definition, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return simpleJob.Definition{}, fmt.Errorf("failed to marshal job input json: %w", err)
	}

	strData := strings.TrimSpace(string(data))

	encryptedData, err := r.encrypter.Encrypt(strData)
	if err != nil {
		return simpleJob.Definition{}, fmt.Errorf("failed to encrypt job input: %w", err)
	}

	return simpleJob.Definition{
		Name:       jobUID,
		Type:       JobType,
		MaxRetries: importJobMaxRetries,
		Timeout:    importJobMaxDuration,
		Data:       base64.StdEncoding.EncodeToString(encryptedData),
	}, nil
}

func (r *Repository) getJobInput(data string) (Input, error) {
	encrypted, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return Input{}, fmt.Errorf("failed to base64 decode job input: %w", err)
	}

	decrypted, err := r.encrypter.Decrypt(encrypted)
	if err != nil {
		return Input{}, fmt.Errorf("failed to decrypt job input: %w", err)
	}

	var input Input

	err = json.NewDecoder(strings.NewReader(decrypted)).Decode(&input)
	if err != nil {
		return Input{}, fmt.Errorf("failed to unmarshal job input json: %w", err)
	}

	return input, nil
}

var _ pkgJobs.Handler = (*Repository)(nil)

type Input struct {
	RepoID     int64  `json:"repoID"`
	GitUser    string `json:"gitUser"`
	GitPass    string `json:"gitPass"`
	CloneURL   string `json:"cloneURL"`
	Mirror     bool   `json:"mirror"`
	NoticeLang string `json:"noticeLang"`
}

const JobType = "repository_import"

func (r *Repository) Register(executor *pkgJobs.Executor) error {
	return executor.Register(JobType, r)
}

func (r *Repository) Handle(ctx context.Context, data string, reportFn pkgJobs.ProgressReporter) (string, error) {
	input, err := r.getJobInput(data)
	if err != nil {
		return "", err
	}

	if input.CloneURL == "" {
		return "", errors.New("missing git repository clone URL")
	}

	repoURL, err := url.Parse(input.CloneURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse git clone URL: %w", err)
	}

	repoURL.User = url.UserPassword(input.GitUser, input.GitPass)
	cloneURLWithAuth := repoURL.String()

	repo, err := r.repoStore.GetByID(ctx, uint(input.RepoID))
	if err != nil {
		return "", fmt.Errorf("failed to find repo by id: %w", err)
	}

	if repo.Status != constant.RepoStateImporting {
		return "", fmt.Errorf("repository %s is not being imported", repo.Identifier)
	}

	log := logger.GetLogger(ctx).With(
		"job.Name", JobIDFromRepoID(input.RepoID),
		"job.Type", JobType,
		"repo.id", int64(repo.ID),
		"repo.path", repo.Path,
	)

	log.Info("create git repository")

	accountInfo := auth.Session{User: &principal.User{
		Account: authn.HeaderAppAdmin,
		Email:   authn.HeaderAdminEmail,
	}}
	gitUID, err := r.createGitRepository(ctx, accountInfo, repo.ID)
	if err != nil {
		return "", fmt.Errorf("failed to create empty git repository: %w", err)
	}

	log.Infof("successfully created git repository with git_uid '%s'", gitUID)

	err = func() error {
		updated, err := r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
			if repo.Status != constant.RepoStateImporting {
				return errors.New("repository has already finished importing")
			}
			repo.GitUID = gitUID
			repo.Mirror = input.Mirror
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update repository prior to the import: %w", err)
		}
		repo = updated

		log.Info("sync repository")

		err = r.syncGitRepository(ctx, &accountInfo, repo, cloneURLWithAuth)
		if err != nil {
			return fmt.Errorf("failed to sync git repository from '%s': %w", input.CloneURL, err)
		}

		log.Infof("successfully synced repository (with default branch: %q)", repo.DefaultBranch)
		log.Info("update repo in DB")

		updated, err = r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
			if repo.Status != constant.RepoStateImporting {
				return errors.New("repository has already finished importing")
			}

			repo.Status = constant.RepoStateActive

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update repository after import: %w", err)
		}
		repo = updated

		r.sendNotice(ctx, repo.ID, repo.Identifier, repo.CreatedBy, input.NoticeLang, false, "", log)
		return nil
	}()
	if err != nil {
		log.Errorw("failed repository import - cleanup git repository", "error", err)
		r.sendNotice(ctx, repo.ID, repo.Identifier, repo.CreatedBy, input.NoticeLang, true, err.Error(), log)

		repo.GitUID = gitUID // make sure to delete the correct directory

		// 同步失败，删除之前创建的空 git 仓库与库表记录。
		// 使用 WithoutCancel 避免上游 ctx 已取消时清理动作被连带跳过。
		cleanupCtx := context.WithoutCancel(ctx)
		if errDel := r.deleteGitRepository(cleanupCtx, &accountInfo, repo); errDel != nil {
			log.Errorw("failed to delete git repository after failed import", "error", errDel)
		}
		if errDel := r.repoStore.HardDelete(cleanupCtx, repo.ID); errDel != nil {
			log.Errorw("failed to delete repo record after failed import", "error", errDel)
		}

		return "", fmt.Errorf("failed to import repository: %w", err)
	}

	err = r.sseStreamer.Publish(ctx, int64(repo.SpaceID), eventModel.SSETypeRepositoryImportCompleted, repo)
	if err != nil {
		log.Errorw("failed to publish import completion SSE", "error", err)
	}

	log.Info("completed repository import")

	return "", nil
}

func (r *Repository) createGitRepository(ctx context.Context, accountInfo auth.Session, repoID uint) (string, error) {
	now := time.Now()

	envVars, err := r.createEnvVars(ctx, accountInfo.User, repoID)
	if err != nil {
		return "", err
	}

	resp, err := r.git.CreateRepository(ctx, &git.CreateRepositoryParams{
		Actor: git.Identity{
			Name:  accountInfo.User.Account,
			Email: accountInfo.User.Email,
		},
		EnvVars:       envVars,
		DefaultBranch: r.defaultBranch,
		Files:         nil,
		Author: &git.Identity{
			Name:  accountInfo.User.Account,
			Email: accountInfo.User.Email,
		},
		AuthorDate: &now,
		Committer: &git.Identity{
			Name:  accountInfo.User.Account,
			Email: accountInfo.User.Email,
		},
		CommitterDate: &now,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create empty git repository: %w", err)
	}

	return resp.UID, nil
}

func (r *Repository) createEnvVars(ctx context.Context, user *principal.User, repoID uint) (map[string]string, error) {
	envVars, err := githook.GenerateEnvironmentVariables(
		ctx,
		r.urlProvider.GetInternalAPIURL(ctx),
		int64(repoID),
		user.Account,
		false,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate git hook environment variables: %w", err)
	}

	return envVars, nil
}

func (r *Repository) syncGitRepository(ctx context.Context,
	accountInfo *auth.Session,
	repo *repoModel.Repo,
	sourceCloneURL string,
) error {
	writeParams, err := r.createRPCWriteParams(ctx, accountInfo, repo)
	if err != nil {
		return err
	}

	_, err = r.git.SyncRepository(ctx, &git.SyncRepositoryParams{
		WriteParams:       writeParams,
		Source:            sourceCloneURL,
		CreateIfNotExists: false,
		RefSpecs:          []string{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"},
		DefaultBranch:     repo.DefaultBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to sync repository: %w", err)
	}

	return nil
}

func (r *Repository) deleteGitRepository(ctx context.Context,
	accountInfo *auth.Session,
	repo *repoModel.Repo,
) error {
	writeParams, err := r.createRPCWriteParams(ctx, accountInfo, repo)
	if err != nil {
		return err
	}

	err = r.git.DeleteRepository(ctx, &git.DeleteRepositoryParams{
		WriteParams: writeParams,
	})
	if err != nil {
		return fmt.Errorf("failed to delete git repository: %w", err)
	}

	return nil
}

func (r *Repository) createRPCWriteParams(ctx context.Context,
	accountInfo *auth.Session,
	repo *repoModel.Repo,
) (git.WriteParams, error) {
	envVars, err := r.createEnvVars(ctx, accountInfo.User, repo.ID)
	if err != nil {
		return git.WriteParams{}, err
	}
	return git.WriteParams{
		Actor: git.Identity{
			Name:  accountInfo.User.Account,
			Email: accountInfo.User.Email,
		},
		RepoUID: repo.GitUID,
		EnvVars: envVars,
		Mirror:  repo.Mirror,
	}, nil
}

func (r *Repository) sendNotice(
	ctx context.Context,
	repoID uint,
	repoName, noticeUser, noticeLang string,
	isFailed bool,
	failReason string,
	log *logger.SugaredMateLogger,
) {
	var notice string
	if isFailed {
		notice = i18n.TWithLocale(noticeLang, string(i18n.RepoImportNotifyFailure), noticeUser, repoName, failReason)
	} else {
		notice = i18n.TWithLocale(noticeLang, string(i18n.RepoImportNotifySuccess), noticeUser, repoName)
	}
	toList := fmt.Sprintf(",%s,", noticeUser)

	noticeData := &notifyStore.Notify{
		ObjectType: "message",
		ObjectID:   0,
		Action:     repoID,
		ToList:     &toList,
		Data:       &notice,
		CreatedBy:  noticeUser,
		Status:     "wait",
	}
	if errNotify := r.notifyStore.Create(context.WithoutCancel(ctx), noticeData); errNotify != nil {
		log.Errorw("failed to create import notify", "error", errNotify)
	}
}
