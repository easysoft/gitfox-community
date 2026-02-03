package common

import "os"

// function used to expand environment variables.
var getenv = os.Getenv

// GetEnv 获取环境变量
func GetEnv(envstr string, fallback ...string) string {
	e := getenv(envstr)
	if e == "" && len(fallback) > 0 {
		e = fallback[0]
	}
	return e
}
