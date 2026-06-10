package svn

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// UpdateParams contains the parameters for svn update.
type UpdateParams struct {
	// WorkingCopyPath is the local filesystem path to the working copy to update.
	WorkingCopyPath string
}

func (p *UpdateParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.WorkingCopyPath == "" {
		return errors.New("working copy path cannot be empty")
	}
	return nil
}

// Update brings the working copy up to the latest revision from the repository.
// This corresponds to: svn update <WorkingCopyPath>
func (s *Service) Update(ctx context.Context, params *UpdateParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	cmd := command.New("update",
		command.WithArg(params.WorkingCopyPath),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrUpdate, "%v", err)
	}
	return nil
}
