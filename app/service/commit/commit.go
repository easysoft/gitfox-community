package commit

import (
	"context"
	"encoding/base64"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/sha"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

type Service struct {
	urlProvider url.Provider
	gitPkg      git.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 CommitSrv 实例
func NewService(urlProvider url.Provider, git git.Interface) *Service {
	return &Service{
		urlProvider: urlProvider,
		gitPkg:      git,
	}
}

func (s *Service) Create(ctx context.Context, req commitModel.ReqForCreate) (*commitModel.RespForCreate, error) {
	var err error
	actions := make([]git.CommitFileAction, len(req.Actions))
	for i, action := range req.Actions {
		var rawPayload []byte
		switch action.Encoding {
		case constant.ContentEncodingTypeBase64:
			rawPayload, err = base64.StdEncoding.DecodeString(action.Payload)
			if err != nil {
				err = errors.WithI18nKey(err, i18n.CommitActionDecodeError)
				return nil, err
			}
		case constant.ContentEncodingTypeUTF8:
			fallthrough
		default:
			// by default we treat content as is
			rawPayload = []byte(action.Payload)
		}

		actions[i] = git.CommitFileAction{
			Action:  git.FileAction(action.Action),
			Path:    action.Path,
			Payload: rawPayload,
			SHA:     sha.SHA{Str: action.SHA},
		}
	}

	// Create internal write params. Note: This will skip the pre-commit protection rules check.
	// writeParams, err := s.CreateRPCInternalWriteParams(ctx, c.urlProvider, session, repo)
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, err
	}
	now := time.Now()
	rpcOut, err := s.gitPkg.CommitFiles(ctx, &git.CommitFilesParams{
		WriteParams:   writeParams,
		Message:       git.CommitMessage(utils.GenerateCommitTitle(req.Message), req.Message),
		Branch:        req.Branch,
		NewBranch:     req.NewBranch,
		Actions:       actions,
		Committer:     &git.Identity{Name: "name", Email: "email"},
		CommitterDate: &now,
		Author:        &git.Identity{Name: req.Author.Name, Email: req.Author.Email},
		AuthorDate:    &now,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return nil, err
	}

	return &commitModel.RespForCreate{SHA: rpcOut.CommitID.String()}, nil
}

func (s *Service) List(ctx context.Context, req commitModel.ReqFilter, repo repoModel.Repo, gitRef string) (commitModel.RespForList, error) {
	// set gitRef to default branch in case an empty reference was provided
	// if gitRef == "" {
	// 	gitRef = repo.DefaultBranch
	// }

	rpcOut, err := s.gitPkg.ListCommits(ctx, &git.ListCommitsParams{
		ReadParams:   git.ReadParams{RepoUID: repo.GitUID},
		GitREF:       gitRef,
		After:        req.After,
		Page:         int32(req.Page),
		Limit:        int32(req.Limit),
		Path:         req.Path,
		Since:        req.Since,
		Until:        req.Until,
		IncludeStats: req.IncludeStats,
		Regex:        true,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return commitModel.RespForList{}, err
	}

	commits := make([]commitModel.Commit, len(rpcOut.Commits))
	for i := range rpcOut.Commits {
		var commit *commitModel.Commit
		commit, err = utils.MapCommit(&rpcOut.Commits[i])
		if err != nil {
			err = errors.WithI18nKey(err, i18n.GitMapOutputError)
			return commitModel.RespForList{}, err
		}
		commits[i] = *commit
	}
	renameDetailList := make([]commitModel.RenameDetails, len(rpcOut.RenameDetails))
	for i := range rpcOut.RenameDetails {
		renameDetails := utils.MapRenameDetails(rpcOut.RenameDetails[i])
		if renameDetails == nil {
			err = errors.WithI18nKey(err, i18n.GitMapOutputError)
			return commitModel.RespForList{}, err
		}
		renameDetailList[i] = *renameDetails
	}
	return commitModel.RespForList{
		Commits:       commits,
		RenameDetails: renameDetailList,
		TotalCommits:  rpcOut.TotalCommits,
	}, nil
}
