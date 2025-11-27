package git

import (
	"bytes"
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
)

type Interface interface {
	GitServicePack(ctx context.Context, options api.ServicePackOptions) error
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

func (s *Service) GitServicePack(ctx context.Context, options api.ServicePackOptions) error {
	// setup read/writeparams depending on whether it's a write operation
	if options.IsWrite {
		// var writeParams git.WriteParams
		// writeParams, err = controller.CreateRPCExternalWriteParams(ctx, c.urlProvider, session, repo)
		// if err != nil {
		// 	return fmt.Errorf("failed to create RPC write params: %w", err)
		// }
		// params.WriteParams = &writeParams
	} else {
		// readParams := git.CreateReadParams(repo)
		// params.ReadParams = &readParams
	}

	// if err = s.pkgGit.ServicePack(ctx, params); err != nil {
	// 	return fmt.Errorf("failed service pack operation %q  on git: %w", options.Service, err)
	// }

	return nil
}
