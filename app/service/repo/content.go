package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

func (s *Service) GetContent(ctx context.Context, session *auth.Session, req *repoModel.GetContentReq) (*repoModel.GetContentOutput, error) {
	// GetRepoById todo
	repo, err := s.GetRepoCheckAccess(ctx, session, "", []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		return nil, err
	}

	if req.GitRef == "" {
		req.GitRef = repo.DefaultBranch
	}

	// create read params once
	// readParams := git.CreateReadParams(repo)
	// treeNodeOutput, err := s.gitPkgSvc.GetTreeNode(ctx, &git.GetTreeNodeParams{
	// 	ReadParams:          readParams,
	// 	GitREF:              req.GitRef,
	// 	Path:                req.RepoPath,
	// 	IncludeLatestCommit: req.IncludeLatestCommit,
	// })

	return nil, nil
}
