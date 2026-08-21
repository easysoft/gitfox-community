package common

import (
	"fmt"
)

var (
	Version       string
	BuildDate     string
	GitCommitHash string
	GitBranch     string
)

// VersionInfo 是 `gitfox version` 命令对外暴露的结构化版本信息。
//
// 该结构体被两处共用：
//  1. version 命令在 --json 模式下序列化输出
//  2. upgrade 命令解析其它 gitfox 二进制的版本信息
//
// 字段未填充时（编译注入失败）返回兜底字符串而非空串，避免下游做 nil 判断。
type VersionInfo struct {
	Version       string `json:"version"`
	BuildDate     string `json:"buildDate"`
	GitBranch     string `json:"gitBranch"`
	GitCommitHash string `json:"gitCommitHash"`
}

// GetVersionInfo 返回当前进程的版本信息，未注入字段会用 "unknown" / "0.0.1" 兜底。
func GetVersionInfo() VersionInfo {
	v := Version
	if v == "" {
		v = "0.0.1"
	}
	bd := BuildDate
	if bd == "" {
		bd = "unknown"
	}
	br := GitBranch
	if br == "" {
		br = "unknown"
	}
	gc := GitCommitHash
	if gc == "" {
		gc = "unknown"
	}
	return VersionInfo{
		Version:       v,
		BuildDate:     bd,
		GitBranch:     br,
		GitCommitHash: gc,
	}
}

// GetVersionWithHash returns the version string with the git commit hash and build date.
func GetVersionWithHash() string {
	return fmt.Sprintf("%s-%s-%s", Version, GitCommitHash, BuildDate)
}
