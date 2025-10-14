package pkg_errors

import (
	"runtime"
)

func GetFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	return "funcName:" + runtime.FuncForPC(pc[0]).Name() + " "
}

const (
	NoType ErrorType = iota
	BadRequest
	NotFound
	DBError
	SystemErr
)

const defaultI18nKey = "error.default"
