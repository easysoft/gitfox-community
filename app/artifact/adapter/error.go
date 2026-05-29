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
