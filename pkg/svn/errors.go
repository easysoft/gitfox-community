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
	ErrLog                    = errors.New("failed to get svn log")
	ErrDiff                   = errors.New("failed to get svn diff")
	ErrLs                     = errors.New("failed to get svn ls")
	ErrCat                    = errors.New("failed to get svn cat")
	ErrBlame                  = errors.New("failed to get svn blame")
	ErrExport                 = errors.New("failed to export svn repository")
	ErrNoParamsProvided       = errors.New("params not provided")
)
