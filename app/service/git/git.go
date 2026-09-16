package git

import (
	"bytes"
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/githook"
	"github.com/easysoft/gitfox-community/app/model/auth"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/utils/url"
)

type Interface interface {
	GitServicePack(ctx context.Context, session *auth.Session, repo *repoModel.Repo, options api.ServicePackOptions) error
	GetInfoRefs(ctx context.Context, repoUID string, service string, gitProtocol string) ([]byte, error)
}

type Service struct {
	git         git.Interface
	urlProvider url.Provider
}

func NewGitService(gitPkg git.Interface, urlProvider url.Provider) Interface {
	return &Service{
		git:         gitPkg,
		urlProvider: urlProvider,
	}
}

func (s *Service) GetInfoRefs(ctx context.Context, repoUID string, service string, gitProtocol string) ([]byte, error) {
	buf := &bytes.Buffer{}

	params := &git.InfoRefsParams{
		ReadParams: git.ReadParams{
			RepoUID: repoUID,
		},
		Service:     service,
		GitProtocol: gitProtocol,
	}

	if err := s.git.GetInfoRefs(ctx, buf, params); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *Service) GitServicePack(ctx context.Context, session *auth.Session, repo *repoModel.Repo, options api.ServicePackOptions) error {
	var params *git.ServicePackParams

	// setup read/writeparams depending on whether it's a write operation
	if options.IsWrite {
		// Generate git hook environment variables for write operations
		envVars, err := githook.GenerateEnvironmentVariables(
			ctx,
			s.urlProvider.GetInternalAPIURL(ctx),
			int64(repo.ID),
			session.User.Account,
			false, // disabled
			false, // internal (this is an external SSH push)
		)
		if err != nil {
			return fmt.Errorf("failed to generate git hook environment variables: %w", err)
		}

		params = &git.ServicePackParams{
			WriteParams: &git.WriteParams{
				Actor: git.Identity{
					Name:  session.User.Account,
					Email: session.User.Email,
				},
				RepoUID: repo.GitUID,
				EnvVars: envVars,
			},
			ServicePackOptions: options,
		}
	} else {
		params = &git.ServicePackParams{
			ReadParams:         &git.ReadParams{RepoUID: repo.GitUID},
			ServicePackOptions: options,
		}
	}

	if err := s.git.ServicePack(ctx, params); err != nil {
		return err
	}

	return nil
}
