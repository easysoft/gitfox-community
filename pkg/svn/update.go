package svn

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// UpdateParams contains the parameters for svn update.
type UpdateParams struct {
	// SvnUID identifies the working copy to update.
	// The working copy path is resolved as {svnRoot}/repos/{svnUID}-wc.
	SvnUID string
}

func (p *UpdateParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Update brings the working copy up to the latest revision from the repository.
// This corresponds to: svn update <workingCopyPath>
func (s *Service) Update(ctx context.Context, params *UpdateParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	cmd := command.New("update",
		command.WithArg(s.workingCopyPath(params.SvnUID)),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrUpdate, "%v", err)
	}
	return nil
}
