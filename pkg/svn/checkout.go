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
	"fmt"
	"os"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn/command"
)

// CheckoutParams contains the parameters for svn checkout.
type CheckoutParams struct {
	// SourceURL is the URL of the SVN repository to check out.
	// e.g. "file:///data/.../demo" or "https://repo.oop.cc/svn/repo/..."
	SourceURL string

	// TargetPath is the local filesystem path where the working copy will be created.
	TargetPath string

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *CheckoutParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SourceURL == "" {
		return errors.New("source url cannot be empty")
	}
	if p.TargetPath == "" {
		return errors.New("target path cannot be empty")
	}
	return nil
}

// Checkout creates a working copy from the source repository at the target path.
// This corresponds to: svn checkout <SourceURL> <TargetPath> --username ... --password ...
func (s *Service) Checkout(ctx context.Context, params *CheckoutParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(params.TargetPath, 0o700); err != nil {
		return errors.Wrapf(ErrCheckout, "failed to create target directory %s: %v", params.TargetPath, err)
	}

	cmd := command.New("checkout",
		command.WithArg(params.SourceURL),
		command.WithArg(params.TargetPath),
		withCheckoutAuth(params.Username, params.Password),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrCheckout, "%v", err)
	}
	return nil
}

// CheckoutFromRepo creates a working copy from a local SVN repository at the given repo path.
// This is a convenience wrapper that constructs the file:// URL from the repo path.
func (s *Service) CheckoutFromRepo(ctx context.Context, repoPath, targetPath string) error {
	return s.Checkout(ctx, &CheckoutParams{
		SourceURL:  fmt.Sprintf("file://%s", repoPath),
		TargetPath: targetPath,
	})
}

// withCheckoutAuth adds --username and --password flags when set.
func withCheckoutAuth(username, password string) command.CmdOptionFunc {
	return func(c *command.Command) {
		if username != "" {
			c.Flags = append(c.Flags, "--username", username)
		}
		if password != "" {
			c.Flags = append(c.Flags, "--password", password)
		}
	}
}
