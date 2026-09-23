// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mengyutong <mengyutong@chandao.com>

package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"
	providerModel "github.com/easysoft/gitfox-community/app/model/provider"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

func (r *Repository) Migrate(ctx context.Context) error {
	log := logger.GetLogger(ctx)

	// 先查询数据库，找出迁移过来的代码库列表
	migrateData, err := r.repoStore.GetMigrate(ctx)
	if err != nil {
		return err
	}
	if len(migrateData) == 0 {
		log.Infow("startup repo migrate: no candidate repos found")
		return nil
	}
	log.Infow("startup repo migrate: loaded candidates", "count", len(migrateData))

	// 过滤出 providerId 列表
	providerIDs := make([]uint, 0, len(migrateData))
	providerSet := make(map[uint]struct{}, len(migrateData))
	for _, repo := range migrateData {
		if _, exists := providerSet[repo.ProviderID]; exists {
			continue
		}
		providerSet[repo.ProviderID] = struct{}{}
		providerIDs = append(providerIDs, repo.ProviderID)
	}
	if len(providerIDs) == 0 {
		log.Infow("startup repo migrate: no provider IDs extracted from candidates")
		return nil
	}

	// 查询代码库对应的 provider 信息列表
	providers, err := r.providerStore.GetByIDs(ctx, providerIDs)
	if err != nil {
		return err
	}
	log.Infow("startup repo migrate: providers loaded", "count", len(providers))

	noticeLang := i18n.GetLocaleFromContext(ctx)
	total := 0
	failed := 0
	scheduled := 0

	for _, repo := range migrateData {
		total++
		jobName := JobIDFromRepoID(int64(repo.ID))
		log.Infow("startup repo migrate: evaluating repo",
			"repoID", repo.ID,
			"providerID", repo.ProviderID,
			"status", repo.Status,
			"scmType", repo.SCMType,
			"mirror", repo.Mirror,
			"gitUID", repo.GitUID,
			"jobName", jobName,
		)

		providerInfo, ok := providers[repo.ProviderID]
		if !ok {
			failed++
			log.Warnw("startup repo migrate: provider not found", "repoID", repo.ID, "providerID", repo.ProviderID, "jobName", jobName)
			continue
		}

		repoProvider, err := buildProviderFromRepo(repo, providerInfo)
		if err != nil {
			failed++
			log.Errorw("startup repo migrate: failed to build provider payload", "repoID", repo.ID, "providerID", repo.ProviderID, "jobName", jobName, "error", err)
			continue
		}
		log.Infow("startup repo migrate: provider payload prepared",
			"repoID", repo.ID,
			"providerType", repoProvider.Type,
			"providerHost", repoProvider.Host,
			"slug", repoProvider.Slug,
			"jobName", jobName,
		)

		// 调用 LoadRepositoryFromProvider 获取可用 cloneURL
		remoteRepository, provider, err := r.LoadRepositoryFromProvider(ctx, repoProvider, repoProvider.Slug)
		if err != nil {
			failed++
			log.Errorw("startup repo migrate: failed to probe remote repository", "repoID", repo.ID, "providerID", repo.ProviderID, "slug", repoProvider.Slug, "jobName", jobName, "error", err)
			continue
		}
		log.Infow("startup repo migrate: remote repository probe succeeded",
			"repoID", repo.ID,
			"providerID", repo.ProviderID,
			"cloneURL", remoteRepository.CloneURL,
			"jobName", jobName,
		)

		updatedRepo := repo
		if repo.Status != constant.RepoStateImporting {
			updatedRepo, err = r.repoStore.UpdateOptLock(ctx, repo, func(repo *repoModel.Repo) error {
				repo.Status = constant.RepoStateImporting
				return nil
			})
			if err != nil {
				failed++
				log.Errorw("startup repo migrate: failed to set status importing before scheduling", "repoID", repo.ID, "providerID", repo.ProviderID, "jobName", jobName, "error", err)
				continue
			}
		}

		if err := r.Run(ctx,
			provider,
			updatedRepo,
			remoteRepository.CloneURL,
			updatedRepo.Mirror,
			noticeLang,
		); err != nil {
			failed++
			log.Errorw("startup repo migrate: failed to schedule importer job", "repoID", repo.ID, "providerID", repo.ProviderID, "jobName", jobName, "error", err)
			continue
		}
		scheduled++
		log.Infow("startup repo migrate: importer job scheduled",
			"repoID", repo.ID,
			"providerID", repo.ProviderID,
			"jobName", jobName,
			"statusAfterUpdate", updatedRepo.Status,
		)
	}

	log.Infow("startup repo migrate: summary", "total", total, "scheduled", scheduled, "failed", failed)

	if failed == total {
		return fmt.Errorf("migrate scheduling completed all failed")
	}

	return nil
}

func buildProviderFromRepo(repo *repoModel.Repo, provider *providerModel.Provider) (repoModel.Provider, error) {
	repoProvider := repoModel.Provider{
		Type:  constant.RepoProviderType(provider.Type),
		ID:    provider.ID,
		Host:  provider.URL,
		Token: provider.Token,
	}

	if repo.SCMType == constant.RepoScmTypeSvn {
		svnConnectInfo := repoModel.SVNConnectInfo{}
		if err := json.Unmarshal(repo.Connector, &svnConnectInfo); err != nil {
			return repoModel.Provider{}, fmt.Errorf("failed to unmarshal SVN connect info: %w", err)
		}

		passwordBytes, err := base64.StdEncoding.DecodeString(svnConnectInfo.Password)
		if err != nil {
			return repoModel.Provider{}, fmt.Errorf("failed to decode SVN password: %w", err)
		}

		repoProvider.Slug = svnConnectInfo.Slug
		repoProvider.Username = svnConnectInfo.User
		repoProvider.Password = string(passwordBytes)
		return repoProvider, nil
	}

	gitConnectInfo := repoModel.GitConnectInfo{}
	if err := json.Unmarshal(repo.Connector, &gitConnectInfo); err != nil {
		return repoModel.Provider{}, fmt.Errorf("failed to unmarshal Git connect info: %w", err)
	}

	repoProvider.Slug = gitConnectInfo.Slug
	repoProvider.ProjectID = gitConnectInfo.ProjectID
	return repoProvider, nil
}
