package svn

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

var (
	ErrSvnBinaryNotFound      = errors.New("svn binary not found")
	ErrSvnVersionNotSupported = errors.New("svn version not supported")
	ErrFailedToGetVersion     = errors.New("failed to get svn version")
	ErrCreateRepository       = errors.New("failed to create svn repository")
	ErrSyncInit               = errors.New("failed to initialize svnsync")
	ErrSync                   = errors.New("failed to sync svnsync")
	ErrSyncCopyRevProps       = errors.New("failed to copy svnsync revprops")
	ErrCheckout               = errors.New("failed to checkout svn working copy")
	ErrUpdate                 = errors.New("failed to update svn working copy")
	ErrRemoveRepository       = errors.New("failed to remove svn repository")
	ErrRemoveWorkingCopy      = errors.New("failed to remove svn working copy")
	ErrInfo                   = errors.New("failed to get svn info")
	ErrNoParamsProvided       = errors.New("params not provided")
)
