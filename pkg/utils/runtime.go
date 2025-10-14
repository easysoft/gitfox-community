package utils

import (
	"runtime"
	"strings"
)

func GetCallerName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "Unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "Unknown"
	}

	fullName := fn.Name()
	if lastDot := strings.LastIndex(fullName, "."); lastDot != -1 {
		return fullName[lastDot+1:]
	}

	return fullName
}
