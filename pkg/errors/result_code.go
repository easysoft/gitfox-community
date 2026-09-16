package errors

import (
	"net/http"

	"github.com/easysoft/gitfox-community/pkg/i18n"
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
	Conflict
	PreconditionFailed
)

const defaultI18nKey = i18n.ErrorDefault

const defaultHttpStatus = http.StatusOK
