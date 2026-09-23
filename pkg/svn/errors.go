// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package svn

import (
	"github.com/easysoft/gitfox-community/pkg/errors"
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
