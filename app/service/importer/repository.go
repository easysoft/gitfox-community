package importer

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/githook"
	artifactModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	simpleJob "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/encrypt"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	notifyStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/notify"
	providerStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/provider"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	webhookStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	svnPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn"
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
	svn           svnPkg.Interface
	repoStore     repoStore.Interface
	providerStore providerStore.Interface
	notifyStore   notifyStore.Interface
	encrypter     encrypt.Encrypter
	scheduler     *jobs.Scheduler
	sseStreamer   sse.Streamer
	devopsConfig  config.DevopsConfig
	webhookStore  webhookStore.Interface
	artSvc        artifactSvc.Interface
}

var _ Interface = (*Repository)(nil)

// NewRepoImporterSvc 创建 RepoService 实例
func NewRepoImporterSvc(
	defaultBranch string,
	urlProvider urlUtil.Provider,
	git git.Interface,
	svn svnPkg.Interface,
	repoStore repoStore.Interface,
	providerStore providerStore.Interface,
	notifyStore notifyStore.Interface,
	encrypter encrypt.Encrypter,
	scheduler *jobs.Scheduler,
	sseStreamer sse.Streamer,
	devopsConfig config.DevopsConfig,
	webhookStore webhookStore.Interface,
	artSvc artifactSvc.Interface,
) *Repository {
	return &Repository{
		defaultBranch: defaultBranch,
		urlProvider:   urlProvider,
		git:           git,
		svn:           svn,
		repoStore:     repoStore,
		notifyStore:   notifyStore,
		providerStore: providerStore,
		encrypter:     encrypter,
		scheduler:     scheduler,
		sseStreamer:   sseStreamer,
		devopsConfig:  devopsConfig,
		webhookStore:  webhookStore,
		artSvc:        artSvc,
	}
}

func (r *Repository) Create(ctx context.Context, session *auth.Session, req *repoModel.Repo) (*repoModel.Repo, error) {
	repo := &repoStore.Repository{
		SpaceID:       req.SpaceID,
		Identifier:    req.Identifier,
		GitUID:        req.GitUID,
		Desc:          req.Desc,
		SCMType:       req.SCMType,
		Product:       req.Product,
		CreatedBy:     session.User.Account,
		DefaultBranch: req.DefaultBranch,
		Status:        req.Status,
		ACL:           req.ACL,
		Mirror:        database.BoolToUint8(req.Mirror),
		ProviderID:    req.ProviderID,
		Connector:     req.Connector,
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

	if provider.Type == constant.RepoProviderTypeSVN {
		cloneURL := fmt.Sprintf("%s/%s", provider.Host, provider.Slug)

		// 保证远端代码库可用
		if _, err := r.svn.InfoRevision(ctx, cloneURL, provider.Username, provider.Password); err != nil {
			return RepositoryInfo{}, provider, errors.WithI18nKey(
				errors.BadRequest.Wrap(err, "svn repository not found or unreachable"),
				i18n.RepoImportSvnNotFoundOrUnreachable,
			)
		}

		return RepositoryInfo{
			CloneURL: cloneURL,
		}, provider, nil
	}

	if slug == "" {
		return RepositoryInfo{}, provider, errors.WithI18nKey(
			errors.BadRequest.New("provider repository identifier is missing"),
			i18n.RepoProviderRepoIdentifierMissing,
		)
	}

	scmClient, err := getScmClientWithTransport(provider, slug)
	if err != nil {
		return RepositoryInfo{}, provider, errors.WithI18nKey(
			errors.BadRequest.Wrap(err, "could not create client"),
			i18n.RepoProviderScmClientCreateError,
		)
	}

	// Augment user information if it's not provided for certain vendors.
	if provider.Token != "" && provider.Username == "" {
		user, _, err := scmClient.Users.Find(ctx)
		if err != nil {
			return RepositoryInfo{}, provider, errors.WithI18nKey(
				fmt.Errorf("could not find user: %s", err),
				i18n.RepoProviderUserFindError,
			)
		}
		provider.Username = user.Login
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
	// 这里的 Input 需要根据 provider 类型进行适当调整
	auth := provider.Token
	if provider.Type == constant.RepoProviderTypeSVN {
		auth = provider.Password
	}

	jobDef, err := r.getJobDef(JobIDFromRepoID(int64(repo.ID)), Input{
		ProviderType: provider.Type,
		RepoID:       int64(repo.ID),
		User:         provider.Username,
		Auth:         auth,
		CloneURL:     cloneURL,
		NoticeLang:   noticeLang,
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
	ProviderType constant.RepoProviderType `json:"providerType"`
	RepoID       int64                     `json:"repoID"`
	User         string                    `json:"user"`
	Auth         string                    `json:"auth"`
	CloneURL     string                    `json:"cloneURL"`
	NoticeLang   string                    `json:"noticeLang"`
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

	if input.ProviderType == constant.RepoProviderTypeGogs {
		repoURL.User = url.User(input.Auth)
	} else {
		repoURL.User = url.UserPassword(input.User, input.Auth)
	}
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

	if input.ProviderType == constant.RepoProviderTypeSVN {
		repo, err = r.firstSyncSvnRepo(ctx, repo, input, log)
		if err != nil {
			return "", err
		}
	} else {
		repo, err = r.firstSyncGitRepo(ctx, repo, cloneURLWithAuth, input, log)
		if err != nil {
			return "", err
		}
	}

	log.Info("completed repository import")
	if repo.Mirror {
		err = r.CreateMirrorSyncCronJob(ctx, input.RepoID, log)
		if err != nil {
			return "", fmt.Errorf("completed repository import, but failed to create mirror sync cron job: %w", err)
		}
	} else {
		if err = r.createDefaultWebhook(ctx, repo, log); err != nil {
			return "", fmt.Errorf("completed repository import, but failed to create webhook: %w", err)
		}
		if err = r.createDefaultArtifactLib(ctx, repo, log); err != nil {
			return "", fmt.Errorf("completed repository import, but failed to create default artifact lib: %w", err)
		}
	}

	return "", nil
}

// createDefaultWebhook 创建默认 zentao webhook，如已存在同名 webhook 则跳过。
func (r *Repository) createDefaultWebhook(ctx context.Context, repo *repoModel.Repo, log *logger.SugaredMateLogger) error {
	displayName := fmt.Sprintf("zentao-%d-%s", repo.ID, time.Now().Format("20060102"))
	_, err := r.webhookStore.GetByRepoIDAndDisplayName(ctx, int64(repo.ID), displayName)
	if err == nil {
		log.Infof("webhook with display name %s already exists, skipping creation", displayName)
		return nil
	}
	if !errors.Is(err, errors.ErrResourceNotFound) {
		log.Errorf("failed to check webhook existence: %v", err)
		return err
	}

	wh := &webhookStore.Webhook{
		RepoID:      int64(repo.ID),
		DisplayName: displayName,
		URL:         fmt.Sprintf("%s/gitfox/webhook?repoID=%d", r.devopsConfig.URL, repo.ID),
		IsSSL:       0,
		Internal:    0,
		AuthMethod:  constant.DefaultAuthMethod,
		AuthHeader:  constant.DefaultAuthHeader,
		Enabled:     1,
		CreatedBy:   repo.CreatedBy,
		EditedBy:    repo.CreatedBy,
	}
	if _, err = r.webhookStore.Create(ctx, wh); err != nil {
		log.Errorf("failed to create webhook: %v", err)
		return err
	}
	return nil
}

// createDefaultArtifactLib 创建默认制品库（空库），与普通建仓时自动创建的默认库保持一致。
// CreateLib 非幂等，同 scope 下已存在同名库时返回 ArtifactLibNameDuplicate，视为已创建跳过。
func (r *Repository) createDefaultArtifactLib(ctx context.Context, repo *repoModel.Repo, log *logger.SugaredMateLogger) error {
	// 后台任务 ctx 无鉴权会话，CreateLib 内部会从 ctx 读取 session 写入 CreatedBy/EditedBy，
	// 这里以应用管理员身份注入，与 firstSyncGitRepo 中 accountInfo 的口径一致。
	ctx = req.SetAuthSession(ctx, &auth.Session{
		User: &principal.User{
			Account: authn.HeaderAppAdmin,
			Email:   authn.HeaderAdminEmail,
		},
	})

	_, err := r.artSvc.CreateLib(ctx, &artifactModel.CreateArtifactLibReq{
		SpaceRepoReq: artifactModel.SpaceRepoReq{
			SpaceID: int64(repo.SpaceID),
			RepoID:  int64(repo.ID),
		},
		Name:   artifactModel.DefaultArtifactLibName,
		Code:   artifactModel.DefaultArtifactLibCode,
		Format: string(artifactModel.ArtifactFileFormat),
	})
	if err != nil {
		if errors.GetI18nKey(err) == i18n.ArtifactLibNameDuplicate {
			log.Infof("default artifact lib already exists, skipping creation")
			return nil
		}
		log.Errorf("failed to create default artifact lib: %v", err)
		return err
	}
	return nil
}

// firstSyncGitRepo 创建空 git 仓库并同步仓库内容，成功后更新库表状态为 Active。
// 同步失败时会自动清理已创建的 git 仓库和库表记录。
func (r *Repository) firstSyncGitRepo(
	ctx context.Context,
	repo *repoModel.Repo,
	cloneURLWithAuth string,
	input Input,
	log *logger.SugaredMateLogger,
) (*repoModel.Repo, error) {
	log.Info("create git repository")

	accountInfo := auth.Session{User: &principal.User{
		Account: authn.HeaderAppAdmin,
		Email:   authn.HeaderAdminEmail,
	}}

	oldGitUID := repo.GitUID
	gitUID, err := r.createGitRepository(ctx, accountInfo, repo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create empty git repository: %w", err)
	}

	log.Infof("successfully created git repository with git_uid '%s'", gitUID)

	err = func() error {
		updated, err := r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
			if repo.Status != constant.RepoStateImporting {
				return errors.New("repository has already finished importing")
			}
			repo.GitUID = gitUID
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

		// 如果是迁移的仓库，不删除库表记录，将 gitUID 置为导入时的标记
		if strings.HasPrefix(oldGitUID, constant.MigrateRepoGitUID) {
			_, err = r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
				repo.GitUID = oldGitUID
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update repository after import: %w", err)
			}

			return nil, fmt.Errorf("failed to import repository: %w", err)
		}

		if errDel := r.repoStore.HardDelete(cleanupCtx, repo.ID); errDel != nil {
			log.Errorw("failed to delete repo record after failed import", "error", errDel)
		}

		return nil, fmt.Errorf("failed to import repository: %w", err)
	}

	return repo, nil
}

// firstSyncSvnRepo 创建本地 SVN 仓库并通过 svnsync 同步远程仓库内容，
// 成功后更新库表状态为 Active。同步失败时会自动清理已创建的 SVN 仓库和库表记录。
func (r *Repository) firstSyncSvnRepo(
	ctx context.Context,
	repo *repoModel.Repo,
	input Input,
	log *logger.SugaredMateLogger,
) (*repoModel.Repo, error) {
	log.Info("create svn repository")

	// 构造固定 42 字符的目录名: SHA-256(spaceID-identifier-timestamp)[:31] → base64
	/* #nosec G401 */
	hash := sha256.Sum256([]byte(
		fmt.Sprintf("%d-%s-%d", repo.SpaceID, repo.Identifier, time.Now().UnixMicro()),
	))
	svnUID := base64.RawURLEncoding.EncodeToString(hash[:31]) // 31 bytes → 42 chars

	if err := r.svn.CreateRepository(ctx, svnUID); err != nil {
		return nil, fmt.Errorf("failed to create svn repository: %w", err)
	}

	log.Infof("successfully created svn repository for '%s'", svnUID)

	oldGitUID := repo.GitUID

	// 更新 GitUID 为 svnUID，确保同步失败时能定位到仓库进行清理
	updated, err := r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
		if repo.Status != constant.RepoStateImporting {
			return errors.New("repository has already finished importing")
		}
		repo.GitUID = svnUID
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update repository prior to the import: %w", err)
	}
	repo = updated

	err = func() error {
		log.Info("initialize svnsync")

		if err := r.svn.SyncInit(ctx, &svnPkg.SyncInitParams{
			SvnUID:         svnUID,
			SourceURL:      input.CloneURL,
			SourceUsername: input.User,
			SourcePassword: input.Auth,
		}); err != nil {
			return fmt.Errorf("failed to initialize svnsync: %w", err)
		}

		log.Info("sync svn repository")

		if err := r.svn.Sync(ctx, &svnPkg.SyncParams{
			SvnUID:         svnUID,
			SourceUsername: input.User,
			SourcePassword: input.Auth,
		}); err != nil {
			return fmt.Errorf("failed to sync svn repository from '%s': %w", input.CloneURL, err)
		}

		log.Info("copy revision properties")

		if err := r.svn.SyncCopyRevProps(ctx, &svnPkg.SyncParams{
			SvnUID:         svnUID,
			SourceUsername: input.User,
			SourcePassword: input.Auth,
		}); err != nil {
			return fmt.Errorf("failed to copy svn revision properties: %w", err)
		}

		log.Info("checkout working copy")

		if err := r.svn.Checkout(ctx, &svnPkg.CheckoutParams{
			SvnUID: svnUID,
		}); err != nil {
			return fmt.Errorf("failed to checkout svn working copy: %w", err)
		}

		log.Infof("successfully checked out svn working copy for '%s'", svnUID)

		log.Info("update repo in DB")

		updated, err := r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
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
		log.Errorw("failed repository import - cleanup svn repository", "error", err)
		r.sendNotice(ctx, repo.ID, repo.Identifier, repo.CreatedBy, input.NoticeLang, true, err.Error(), log)

		cleanupCtx := context.WithoutCancel(ctx)
		if errDel := r.svn.RemoveRepository(svnUID); errDel != nil {
			log.Errorw("failed to delete svn repository after failed import", "error", errDel)
		}
		if errDel := r.svn.RemoveWorkingCopy(svnUID); errDel != nil {
			log.Errorw("failed to delete svn working copy after failed import", "error", errDel)
		}

		// 如果是迁移的仓库，不删除库表记录，将 gitUID 置为迁移过来的标记
		if strings.HasPrefix(oldGitUID, constant.MigrateRepoGitUID) {
			_, err = r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
				repo.GitUID = oldGitUID
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update repository after import: %w", err)
			}

			return nil, fmt.Errorf("failed to import repository: %w", err)
		}

		if errDel := r.repoStore.HardDelete(cleanupCtx, repo.ID); errDel != nil {
			log.Errorw("failed to delete repo record after failed import", "error", errDel)
		}

		return nil, fmt.Errorf("failed to import repository: %w", err)
	}

	return repo, nil
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
		notice = i18n.TWithLocale(noticeLang, string(i18n.RepoImportNotifyFailure), repoName, failReason)
	} else {
		notice = i18n.TWithLocale(noticeLang, string(i18n.RepoImportNotifySuccess), repoName)
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

// NewMirrorSyncHandler 创建一个镜像同步任务处理器，供外部注册到 JobScheduler。
func (r *Repository) NewMirrorSyncHandler() *MirrorSyncHandler {
	return NewMirrorSyncHandler(r.git, r.svn, r.repoStore, r.providerStore, r.encrypter)
}
