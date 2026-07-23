//go:build !windows

package service

// windows_svc_stub.go 是 windows_svc.go 的非 Windows 平台桩文件。
// 提供空实现确保 Linux / 其他平台编译通过。

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/quicklyon/kingpin/v2"
)

func checkSudoWindows() error {
	return errors.New("Windows 管理员检测仅在 Windows 平台可用")
}

func registerWindows(_, _ string) error {
	return errors.New("Windows 服务注册仅在 Windows 平台可用")
}

func restartWindows() error {
	return errors.New("Windows 服务重启仅在 Windows 平台可用")
}

func stopWindows() error {
	return errors.New("Windows 服务停止仅在 Windows 平台可用")
}

func uninstallWindows() error {
	return errors.New("Windows 服务卸载仅在 Windows 平台可用")
}

func inspectWindows() ServiceInfo {
	return ServiceInfo{Registered: false}
}

func RegisterServiceRun(_ *kingpin.Application) {
}

func newDockerHTTPClientNpipePath(_ string) (*http.Client, string, error) {
	return nil, "", fmt.Errorf("Windows npipe 仅在 Windows 平台可用")
}

func dockerNpipeExists(_ string) bool {
	return false
}
