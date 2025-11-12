package tag

import (
	"context"
	"time"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	tagModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

type Service struct {
	urlProvider url.Provider
	gitPkg      git.Interface
}

var _ Interface = (*Service)(nil)

// NewService 创建 TagSrv 实例
func NewService(urlProvider url.Provider, git git.Interface) *Service {
	return &Service{
		urlProvider: urlProvider,
		gitPkg:      git,
	}
}

func (s *Service) Create(
	ctx context.Context,
	req tagModel.ReqForCreate,
	session *authModel.Session,
	repo *repoModel.Repo,
) (*tagModel.Tag, error) {

	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return nil, err
	}

	now := time.Now()
	rpcOut, err := s.gitPkg.CreateCommitTag(ctx, &git.CreateCommitTagParams{
		WriteParams: writeParams,
		Name:        req.Name,
		Target:      req.Source,
		Message:     req.Message,
		Tagger:      &git.Identity{Name: session.Principal.UID, Email: session.Principal.Email}, // 从上下文中获取
		TaggerDate:  &now,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return nil, err
	}

	commitTag, err := mapCommitTag(rpcOut.CommitTag)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitMapOutputError)
		return nil, err
	}

	return &commitTag, nil
}

func (s *Service) List(ctx context.Context, req tagModel.ReqFilter, repo repoModel.Repo) ([]tagModel.Tag, error) {

	rpcOut, err := s.gitPkg.ListCommitTags(ctx, &git.ListCommitTagsParams{
		ReadParams:    git.ReadParams{RepoUID: repo.GitUID},
		IncludeCommit: req.IncludeCommit,
		Query:         req.Query,
		Sort:          mapToRPCTagSortOption(req.Sort),
		Order:         utils.MapToRPCSortOrder(req.Order),
		Page:          int32(req.Page),
		PageSize:      int32(req.Size),
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return nil, err
	}

	tags := make([]tagModel.Tag, len(rpcOut.Tags))
	for i := range rpcOut.Tags {
		tags[i], err = mapCommitTag(rpcOut.Tags[i])
		if err != nil {
			err = errors.WithI18nKey(err, i18n.GitExecError)
			return nil, err
		}
	}

	return tags, nil

}

func (s *Service) Delete(ctx context.Context, name string, session *authModel.Session, repo *repoModel.Repo) error {
	// 准备 RPC 参数
	writeParams, err := utils.CreateRPCInternalWriteParams(ctx, s.urlProvider, session, repo)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitCreateExecParamError)
		return err
	}
	err = s.gitPkg.DeleteTag(ctx, &git.DeleteTagParams{
		Name:        name,
		WriteParams: writeParams,
	})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.GitExecError)
		return err
	}
	return nil
}

func mapToRPCTagSortOption(o string) git.TagSortOption {
	switch o {
	case "date":
		return git.TagSortOptionDate
	case "name":
		return git.TagSortOptionName
	default:
		return git.TagSortOptionDefault
	}
}

func mapCommitTag(t git.CommitTag) (tagModel.Tag, error) {
	var commit *commitModel.Commit
	if t.Commit != nil {
		var err error
		commit, err = utils.MapCommit(t.Commit)
		if err != nil {
			return tagModel.Tag{}, err
		}
	}

	var tagger *commitModel.Signature
	if t.Tagger != nil {
		var err error
		tagger, err = utils.MapSignature(t.Tagger)
		if err != nil {
			return tagModel.Tag{}, err
		}
	}

	return tagModel.Tag{
		Name:      t.Name,
		SHA:       t.SHA.String(),
		Message:   t.Message,
		CreatedBy: tagger,
		Commit:    commit,
	}, nil
}
