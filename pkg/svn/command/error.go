// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package command

import (
	"github.com/easysoft/gitfox-community/pkg/errors"
)

var (
	// ErrInvalidArg represent family of errors to report about bad argument used to make a call.
	ErrInvalidArg = errors.New("invalid argument")
)

// NewError creates an error that wraps the original error and includes stderr output.
// The original error (often *exec.ExitError) is preserved in the error chain and can be
// inspected via errors.Is / errors.As.
func NewError(err error, stderr []byte) error {
	if len(stderr) > 0 {
		return errors.Newf("svn: %s: %s", err.Error(), string(stderr))
	}
	return errors.Newf("svn: %s", err.Error())
}
