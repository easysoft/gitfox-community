package errors

import "net/http"

const (
	NoType ErrorType = iota
	BadRequest
	NotFound
	DBError
	SystemErr
)

const defaultI18nKey = "error.default"

const defaultHttpStatus = http.StatusOK
