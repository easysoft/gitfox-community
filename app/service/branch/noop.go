// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package branch

import (
	"context"

	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	branchModel "github.com/easysoft/gitfox-community/app/model/branch"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/gin-gonic/gin"
)

const archiveUnavailableMessage = "branch archive is a paid feature and is not available in this edition"

// NoopArchiveService is the community-edition implementation of IArchiveService.
// The branch archive/activate capability is a paid feature; the paid edition
// binds its own implementation.
type NoopArchiveService struct{}

var _ IArchiveService = (*NoopArchiveService)(nil)

// NewNoopArchiveService creates the community-edition IArchiveService implementation.
func NewNoopArchiveService() *NoopArchiveService {
	return &NoopArchiveService{}
}

// BatchArchive reports that the feature is unavailable for every requested branch.
func (s *NoopArchiveService) BatchArchive(
	_ *gin.Context,
	_ context.Context,
	req branchModel.ReqForArchive,
	_ *authModel.Session,
	_ *repoModel.Repo,
) []branchModel.ArchiveResult {
	results := make([]branchModel.ArchiveResult, 0, len(req.BranchNames))
	for _, name := range req.BranchNames {
		results = append(results, branchModel.ArchiveResult{
			BranchName: name,
			Status:     false,
			Message:    archiveUnavailableMessage,
		})
	}
	return results
}

// BatchActivate reports that the feature is unavailable for every requested branch.
func (s *NoopArchiveService) BatchActivate(
	_ *gin.Context,
	_ context.Context,
	req branchModel.ReqForArchive,
	_ *authModel.Session,
	_ *repoModel.Repo,
) []branchModel.ActivateResult {
	results := make([]branchModel.ActivateResult, 0, len(req.BranchNames))
	for _, name := range req.BranchNames {
		results = append(results, branchModel.ActivateResult{
			BranchName: name,
			Status:     false,
			Message:    archiveUnavailableMessage,
		})
	}
	return results
}
