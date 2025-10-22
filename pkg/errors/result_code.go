package errors

const (
	NoType ErrorType = iota
	BadRequest
	NotFound
	DBError
	SystemErr
)

const defaultI18nKey = "error.default"
