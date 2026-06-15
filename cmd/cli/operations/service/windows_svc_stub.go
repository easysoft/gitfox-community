//go:build !windows

package service

// windows_svc_stub.go 是 windows_svc.go 的非 Windows 平台桩文件。
// 提供空实现确保 Linux / 其他平台编译通过。

import (
	"errors"

	"github.com/quicklyon/kingpin/v2"
)

// checkSudoWindows 非 Windows 平台始终返回 nil（不会被实际调用到）。
func checkSudoWindows() error {
	return errors.New("Windows 管理员检测仅在 Windows 平台可用")
}

// registerWindows 非 Windows 平台不支持服务注册。
func registerWindows(_, _ string) error {
	return errors.New("Windows 服务注册仅在 Windows 平台可用")
}

// restartWindows 非 Windows 平台不支持服务重启。
func restartWindows() error {
	return errors.New("Windows 服务重启仅在 Windows 平台可用")
}

// stopWindows 非 Windows 平台不支持服务停止。
func stopWindows() error {
	return errors.New("Windows 服务停止仅在 Windows 平台可用")
}

// uninstallWindows 非 Windows 平台不支持服务卸载。
func uninstallWindows() error {
	return errors.New("Windows 服务卸载仅在 Windows 平台可用")
}

// inspectWindows 非 Windows 平台返回未注册状态。
func inspectWindows() ServiceInfo {
	return ServiceInfo{Registered: false}
}

// RegisterServiceRun 非 Windows 平台不注册 service run 子命令。
func RegisterServiceRun(_ *kingpin.Application) {
	// 非 Windows 平台无需注册此命令
}
