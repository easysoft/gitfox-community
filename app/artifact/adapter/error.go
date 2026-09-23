// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package adapter

import "errors"

var (
	ErrMissFormField = errors.New("miss required form field")
	ErrMissPathField = errors.New("path field is missing")

	ErrInvalidPackageName       = errors.New("invalid package name")
	ErrInvalidPackageVersion    = errors.New("invalid package version")
	ErrInvalidGroupName         = errors.New("invalid group name")
	ErrInvalidGroupId           = errors.New("invalid group id")
	ErrInvalidPackageContent    = errors.New("invalid package content")
	ErrStorageFileAlreadyExists = errors.New("storage file already exists")
	ErrStorageFileNotChanged    = errors.New("storage file not changed")
)
