package service

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/i18n"
	clii18n "github.com/easysoft/gitfox-community/pkg/i18n/cli"
)

// ServiceInfo 描述系统中已注册的 GitFox 服务的元数据
type ServiceInfo struct {
	Registered bool   // 服务文件是否存在
	ExePath    string // 注册的二进制路径
	WorkDir    string // 注册的工作目录
	Corrupted  bool   // 文件存在但解析失败
}

// InspectService 检查当前系统中是否已注册 GitFox 服务，并提取其元数据。
func InspectService() ServiceInfo {
	switch runtime.GOOS {
	case "linux":
		return inspectLinux()
	case "windows":
		return inspectWindows() // → windows.go
	default:
		return ServiceInfo{}
	}
}

func inspectLinux() ServiceInfo {
	servicePath := "/etc/systemd/system/gitfox.service"
	content, err := os.ReadFile(servicePath)
	if err != nil {
		return ServiceInfo{Registered: false}
	}

	info := ServiceInfo{Registered: true}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ExecStart=") {
			// ExecStart=/path/to/gitfox server prod
			parts := strings.Fields(strings.TrimPrefix(line, "ExecStart="))
			if len(parts) > 0 {
				info.ExePath = parts[0]
			}
		}
		if strings.HasPrefix(line, "WorkingDirectory=") {
			info.WorkDir = strings.TrimPrefix(line, "WorkingDirectory=")
		}
	}

	if info.ExePath == "" {
		info.Corrupted = true
	}
	return info
}

// RestartService 重启已注册的 GitFox 守护进程（不重新生成服务文件）
func RestartService() error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("systemctl", "restart", "gitfox").Run()
	case "windows":
		return restartWindows() // → windows.go
	default:
		return errors.New(clii18n.T(i18n.CliStartErrUnsupportedOs))
	}
}

// validateServicePaths 校验即将写入系统服务文件的路径，防止注入攻击。
// 拒绝包含换行符、分号等可被 systemd 解析器利用的特殊字符。
func validateServicePaths(exePath, workDir string) error {
	dangerousChars := []string{"\n", "\r", ";", "`", "$", "|", "&"}
	for _, c := range dangerousChars {
		if strings.Contains(exePath, c) || strings.Contains(workDir, c) {
			return fmt.Errorf("path contains illegal character %q, aborting service registration for security", c)
		}
	}
	if !filepath.IsAbs(exePath) || !filepath.IsAbs(workDir) {
		return fmt.Errorf("service registration requires absolute paths, got exe=%q workdir=%q", exePath, workDir)
	}
	return nil
}

// RegisterService 根据当前操作系统的不同（Windows / Linux），
// 提取可执行文件物理路径，将其以守护进程方式挂载至目标操作系统的原生后台常驻服务体系中。
func RegisterService(exePath, workDir string) error {
	// 安全校验：防止路径注入
	if err := validateServicePaths(exePath, workDir); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		return registerWindows(exePath, workDir) // → windows.go
	case "linux":
		return registerLinux(exePath, workDir)
	default:
		return errors.New(clii18n.T(i18n.CliStartErrUnsupportedOs))
	}
}

// registerLinux 将程序注册为 standard systemd unit。
func registerLinux(exePath, workDir string) error {
	t := clii18n.T
	// fmt.Println(t(i18n.CliStartStepLinuxService))

	// 1. 前置依赖断言：探测到 systemctl 工具
	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf(t(i18n.CliStartErrSystemctlMissing), workDir)
	}

	// 2. 动态生成符合当前系统的 service definition
	serviceContent := fmt.Sprintf(`[Unit]
Description=GitFox Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s server prod
Restart=on-failure

[Install]
WantedBy=multi-user.target
`, workDir, exePath)

	servicePath := "/etc/systemd/system/gitfox.service"
	// 3. 将 unit 写出至系统底层
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf(t(i18n.CliStartFailed), err)
	}

	// 4. 重载 daemon、设为开机启动、并即刻触发拉起
	_ = exec.Command("systemctl", "daemon-reload").Run()
	_ = exec.Command("systemctl", "enable", "gitfox").Run()
	if err := exec.Command("systemctl", "start", "gitfox").Run(); err != nil {
		return fmt.Errorf(t(i18n.CliStartFailed), err)
	}

	return nil
}
