package git

import (
	"bytes"
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
)

type Interface interface {
	GitServicePack(ctx context.Context, session *auth.Session, repoUID string, options api.ServicePackOptions) error
	GetInfoRefs(ctx context.Context, repoUID string, service string, gitProtocol string) ([]byte, error)
}

type Service struct {
	git git.Interface
}

func NewGitService(gitPkg git.Interface) Interface {
	return &Service{git: gitPkg}
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

func (s *Service) GitServicePack(ctx context.Context, session *auth.Session, repoUID string, options api.ServicePackOptions) error {
	var params *git.ServicePackParams

	// setup read/writeparams depending on whether it's a write operation
	if options.IsWrite {
		params = &git.ServicePackParams{
			WriteParams: &git.WriteParams{
				Actor: git.Identity{
					Name:  session.User.Account,
					Email: session.User.Email,
				},
				RepoUID: repoUID,
			},
			ServicePackOptions: options,
		}
	} else {
		params = &git.ServicePackParams{
			ReadParams:         &git.ReadParams{RepoUID: repoUID},
			ServicePackOptions: options,
		}
	}

	if err := s.git.ServicePack(ctx, params); err != nil {
		return err
	}

	return nil
}
