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
	// ErrAssetDuplicate 表示同名资产已存在且 md5/sha1/sha256 三摘要全等：
	// 视为同一文件重复上传，业务层应直接拒收，不再覆盖、不再自动重命名。
	ErrAssetDuplicate = errors.New("artifact asset already exists with identical content")
)
