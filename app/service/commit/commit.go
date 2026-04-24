package commit

import (
	"context"
	"encoding/base64"
	"io"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
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

func (s *Service) Create(
	ctx context.Context,
	req commitModel.ReqForCreate,
	session *authModel.Session,
	repo *repoModel.Repo,
) (*commitModel.RespForCreate, error) {
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
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	rpcOut, err := s.gitPkg.CommitFiles(ctx, &git.CommitFilesParams{
		WriteParams:   writeParams,
		Message:       git.CommitMessage(utils.GenerateCommitTitle(req.Message), req.Message),
		Branch:        utils.DecodeGitRef(req.Branch),
		NewBranch:     utils.DecodeGitRef(req.NewBranch),
		Actions:       actions,
		Committer:     &git.Identity{Name: session.User.Account, Email: session.User.Email},
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
	gitRef = utils.DecodeGitRef(gitRef)

	if gitRef == "" {
		gitRef = repo.DefaultBranch
	}

	rpcOut, err := s.gitPkg.ListCommits(ctx, &git.ListCommitsParams{
		ReadParams:   git.ReadParams{RepoUID: repo.GitUID},
		GitREF:       gitRef,
		After:        req.After,
		Page:         int32(req.Page),
		Limit:        int32(req.PageSize),
		Path:         req.Path,
		Since:        req.Since,
		Until:        req.Until,
		IncludeStats: req.IncludeStats,
		Committer:    req.Committer,
		Author:       req.Author,
		Regex:        true,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return commitModel.RespForList{}, err
	}

	commits := rpcOut.Commits
	response := make([]commitModel.Commit, len(commits))
	for i := range commits {
		var commit *commitModel.Commit
		commit, err = utils.MapCommit(&commits[i])
		if err != nil {
			err = errors.WithI18nKey(err, i18n.GitMapOutputError)
			return commitModel.RespForList{}, err
		}
		response[i] = *commit
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
		Commits:       response,
		RenameDetails: renameDetailList,
		TotalCommits:  rpcOut.TotalCommits,
	}, nil
}

func (s *Service) Get(ctx context.Context, commitSHA string, repo repoModel.Repo) (*commitModel.Commit, error) {
	rpcOut, err := s.gitPkg.GetCommit(ctx, &git.GetCommitParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		Revision:   commitSHA,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return nil, err
	}

	commit, err := utils.MapCommit(&rpcOut.Commit)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitMapOutputError)
		return nil, err
	}
	return commit, nil
}

func (s *Service) GetDivergence(ctx context.Context, req commitModel.ReqForDivergence, repo repoModel.Repo) ([]commitModel.CommitDivergence, error) {

	// map to rpc params
	options := &git.GetCommitDivergencesParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		MaxCount:   int32(req.MaxDivergence),
		Requests:   make([]git.CommitDivergenceRequest, len(req.Requests)),
	}
	for i := range req.Requests {
		options.Requests[i].From = utils.DecodeGitRef(req.Requests[i].From)
		options.Requests[i].To = utils.DecodeGitRef(req.Requests[i].To)
		// backfil default branch if no 'to' was provided
		if len(options.Requests[i].To) == 0 {
			options.Requests[i].To = repo.DefaultBranch
		}
	}

	rpcOutput, err := s.gitPkg.GetCommitDivergences(ctx, options)
	if err != nil {
		return nil, err
	}

	// map to output type
	divergences := make([]commitModel.CommitDivergence, len(rpcOutput.Divergences))
	for i := range rpcOutput.Divergences {
		divergences[i] = commitModel.CommitDivergence(rpcOutput.Divergences[i])
	}
	return divergences, nil
}

// GetDiff returns the raw git diff for a specific commit in text/plain format.
// The diff is written directly to the provided io.Writer.
func (s *Service) GetDiff(ctx context.Context, commitSHA string, repo repoModel.Repo, w io.Writer) error {
	// Call git package's CommitDiff which writes directly to the writer
	err := s.gitPkg.CommitDiff(ctx, &git.GetCommitParams{
		ReadParams: git.ReadParams{RepoUID: repo.GitUID},
		Revision:   commitSHA,
	}, w)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return err
	}
	return nil
}
