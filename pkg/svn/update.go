// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
