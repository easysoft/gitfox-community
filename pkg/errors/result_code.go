package errors

import (
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

const (
	NoType ErrorType = iota
	BadRequest
	NotFound
	DBError
	GitError
	SystemErr
	InternalError
	InvalidArgument
	ConfigNotFoundError
)

const defaultI18nKey = i18n.ErrorDefault

const defaultHttpStatus = http.StatusOK
