package command

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
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
