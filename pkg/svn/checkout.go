// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"context"
	"os"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/svn/command"
)

// CheckoutParams contains the parameters for svn checkout.
type CheckoutParams struct {
	// SvnUID identifies the SVN repository to check out from.
	// The source URL is constructed as file://{svnRoot}/repos/{svnUID}
	// and the target path as {svnRoot}/repos/{svnUID}-wc.
	SvnUID string

	// Username is the username for authenticating to the repository.
	Username string

	// Password is the password for authenticating to the repository.
	Password string
}

func (p *CheckoutParams) Validate() error {
	if p == nil {
		return ErrNoParamsProvided
	}
	if p.SvnUID == "" {
		return errors.New("svnUID cannot be empty")
	}
	return nil
}

// Checkout creates a working copy from the local SVN repository identified by svnUID.
// The source URL is constructed as file://{svnRoot}/repos/{svnUID}
// and the working copy is placed at {svnRoot}/repos/{svnUID}-wc.
func (s *Service) Checkout(ctx context.Context, params *CheckoutParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	sourceURL := s.FileURL(params.SvnUID)
	targetPath := s.workingCopyPath(params.SvnUID)

	// Ensure parent directory exists.
	if err := os.MkdirAll(targetPath, 0o700); err != nil {
		return errors.Wrapf(ErrCheckout, "failed to create target directory %s: %v", targetPath, err)
	}

	cmd := command.New("checkout",
		command.WithArg(sourceURL),
		command.WithArg(targetPath),
		withCheckoutAuth(params.Username, params.Password),
	)
	if err := cmd.Run(ctx); err != nil {
		return errors.Wrapf(ErrCheckout, "%v", err)
	}
	return nil
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
