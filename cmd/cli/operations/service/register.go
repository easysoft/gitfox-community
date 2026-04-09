package service

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
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
	case "darwin":
		return inspectMac()
	case "windows":
		return inspectWindows()
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

func inspectMac() ServiceInfo {
	plistPath := "/Library/LaunchDaemons/com.gitfox.service.plist"
	content, err := os.ReadFile(plistPath)
	if err != nil {
		return ServiceInfo{Registered: false}
	}

	info := ServiceInfo{Registered: true}

	// 简易提取 ProgramArguments 中的第一个 <string> 和 WorkingDirectory
	exeRe := regexp.MustCompile(`<key>ProgramArguments</key>\s*<array>\s*<string>([^<]+)</string>`)
	wdRe := regexp.MustCompile(`<key>WorkingDirectory</key>\s*<string>([^<]+)</string>`)

	if m := exeRe.FindStringSubmatch(string(content)); len(m) > 1 {
		info.ExePath = m[1]
	}
	if m := wdRe.FindStringSubmatch(string(content)); len(m) > 1 {
		info.WorkDir = m[1]
	}

	if info.ExePath == "" {
		info.Corrupted = true
	}
	return info
}

func inspectWindows() ServiceInfo {
	// 通过 nssm 查询服务参数
	info := ServiceInfo{}
	if err := exec.Command("sc", "query", "gitfox").Run(); err != nil {
		return ServiceInfo{Registered: false}
	}
	info.Registered = true

	// 尝试通过 nssm 获取 Application 路径
	if out, err := exec.Command("nssm", "get", "gitfox", "Application").Output(); err == nil {
		info.ExePath = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("nssm", "get", "gitfox", "AppDirectory").Output(); err == nil {
		info.WorkDir = strings.TrimSpace(string(out))
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
	case "darwin":
		plistPath := "/Library/LaunchDaemons/com.gitfox.service.plist"
		_ = exec.Command("launchctl", "unload", "-w", plistPath).Run()
		return exec.Command("launchctl", "load", "-w", plistPath).Run()
	case "windows":
		_ = exec.Command("nssm", "restart", "gitfox").Run()
		return nil
	default:
		return errors.New(clii18n.T(i18n.CliStartErrUnsupportedOs))
	}
}

// validateServicePaths 校验即将写入系统服务文件的路径，防止注入攻击。
// 拒绝包含换行符、分号等可被 systemd/plist 解析器利用的特殊字符。
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

// RegisterService 根据当前操作系统的不同（Windows / Linux / macOS），
// 提取可执行文件物理路径，将其以守护进程方式挂载至目标操作系统的原生后台常驻服务体系中。
func RegisterService(exePath, workDir string) error {
	// 安全校验：防止路径注入
	if err := validateServicePaths(exePath, workDir); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		return registerWindows(exePath, workDir)
	case "linux":
		return registerLinux(exePath, workDir)
	case "darwin":
		return registerMac(exePath, workDir)
	default:
		return errors.New(clii18n.T(i18n.CliStartErrUnsupportedOs))
	}
}

// registerWindows 将程序注册为 Windows Native Service。
// 注：依赖外部 NSSM 工具包。它会自动处理旧服务的卸载，并注入新的 Application 环境目录与参数。
func registerWindows(exePath, workDir string) error {
	t := clii18n.T
	fmt.Println(t(i18n.CliStartStepWindowsService))

	// 1. 前置依赖断言：NSSM 是否存在于机器的 $PATH 中
	if _, err := exec.LookPath("nssm"); err != nil {
		return errors.New(t(i18n.CliStartErrNssmMissing))
	}

	// 2. 清理旧实例 (防御式编程，允许这里产生 error 不影响后续)
	_ = exec.Command("nssm", "stop", "gitfox").Run()
	_ = exec.Command("nssm", "remove", "gitfox", "confirm").Run()

	// 3. 注册新服务，直接下发 server 端启动命令
	if err := exec.Command("nssm", "install", "gitfox", exePath, "server", "prod").Run(); err != nil {
		return errors.New(t(i18n.CliStartErrRequiresAdmin))
	}

	// 4. 重定向工作目录以确保相对路径（如 db logs）产出正确
	_ = exec.Command("nssm", "set", "gitfox", "AppDirectory", workDir).Run()

	// 5. 扣动扳机启动系统服务
	if err := exec.Command("nssm", "start", "gitfox").Run(); err != nil {
		return fmt.Errorf(t(i18n.CliStartFailed), err)
	}

	return nil
}

// registerLinux 将程序注册为 standard systemd unit。
func registerLinux(exePath, workDir string) error {
	t := clii18n.T
	// fmt.Println(t(i18n.CliStartStepLinuxService))

	// 1. 前置依赖断言：探测到 systemctl 工具
	if _, err := exec.LookPath("systemctl"); err != nil {
		return errors.New(t(i18n.CliStartErrSystemctlMissing))
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

// registerMac 针对 macOS / Darwin 环境，将程序挂载入用户级的 Launchd Daemon。
func registerMac(exePath, workDir string) error {
	t := clii18n.T
	fmt.Println(t(i18n.CliStartStepMacService))

	// 1. 前置依赖断言：探测到 launchctl 工具
	if _, err := exec.LookPath("launchctl"); err != nil {
		return errors.New(t(i18n.CliStartErrLaunchctlMissing))
	}

	// 2. 构建 Apple 标准的 PropertyList XML 配置文件
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.gitfox.service</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>server</string>
        <string>prod</string>
    </array>
    <key>WorkingDirectory</key>
    <string>%s</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>`, exePath, workDir)

	// 3. 将配置抛射进全局守护进程配置目录 (需 sudo 写权限)
	launchDaemonsDir := "/Library/LaunchDaemons"
	if err := os.MkdirAll(launchDaemonsDir, 0755); err != nil {
		return fmt.Errorf(t(i18n.CliStartFailed), err)
	}

	plistPath := filepath.Join(launchDaemonsDir, "com.gitfox.service.plist")
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf(t(i18n.CliStartFailed), err)
	}

	// 4. 通过 launchctl 进行代理挂载及拉起 (-w 强制覆盖 Disabled 状态)
	_ = exec.Command("launchctl", "unload", "-w", plistPath).Run()
	if err := exec.Command("launchctl", "load", "-w", plistPath).Run(); err != nil {
		return fmt.Errorf(t(i18n.CliStartFailed), err)
	}

	return nil
}
