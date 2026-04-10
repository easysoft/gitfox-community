package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	_ "github.com/go-sql-driver/mysql"
)

// CheckSudo 检查当前上下文是否具有超级管理员执行权限 (Root/Sudo/Administrator)。
// 主要由于许多 CLI 的核心命令 (如挂载、停止后台守护进程) 需要穿透系统底层安全限制。
func CheckSudo() error {
	switch runtime.GOOS {
	case "windows":
		return checkSudoWindows()
	default:
		if os.Getuid() != 0 {
			return errors.New(clii18n.T(i18n.CliStartErrRequiresAdmin))
		}
		return nil
	}
}

// CheckGit 检查 git 是否安装并获取版本信息。
// 返回: (是否安装, 解析道德版本号, 错误信息)
func CheckGit() (bool, string, error) {
	t := clii18n.T
	if _, err := exec.LookPath("git"); err != nil {
		return false, "", fmt.Errorf(t(i18n.CliCheckErrNotFound), "git")
	}

	cmd := exec.Command("git", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, "", fmt.Errorf(t(i18n.CliCheckErrExecFailed), "git", err)
	}

	// 标准输出形如: "git version 2.39.2.windows.1"
	outStr := strings.TrimSpace(string(output))
	version := strings.TrimPrefix(outStr, "git version ")
	if version == outStr {
		// 没有匹配到 "git version " 前缀，尝试直接用正则提取数字版本
		re := regexp.MustCompile(`(\d+\.\d+\.\d+(?:\.[-a-zA-Z0-9]+)*)`)
		if matches := re.FindStringSubmatch(outStr); len(matches) > 1 {
			return true, matches[1], nil
		}
		// 如果正则都没提取出来，退化到返回整段字符串
		return true, outStr, nil
	}

	return true, version, nil
}

// CheckDocker 检查 docker 是否安装、daemon 是否正在运行，并获取版本信息。
func CheckDocker() (bool, string, error) {
	t := clii18n.T
	if _, err := exec.LookPath("docker"); err != nil {
		return false, "", fmt.Errorf(t(i18n.CliCheckErrNotFound), "docker")
	}

	cmd := exec.Command("docker", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, "", fmt.Errorf(t(i18n.CliCheckErrExecFailed), "docker", err)
	}

	// 标准输出形如: "Docker version 24.0.5, build ced0996"
	outStr := strings.TrimSpace(string(output))
	re := regexp.MustCompile(`[Vv]ersion\s+(\d+\.\d+\.\d+(?:[-a-zA-Z0-9.]+)*)`)
	matches := re.FindStringSubmatch(outStr)
	version := outStr
	if len(matches) > 1 {
		version = matches[1]
	}

	// 验证 Docker daemon 是否正在运行
	// docker --version 只是客户端本地操作，daemon 停了也能成功
	// docker info 需要和 daemon 通信，可以真正检测 daemon 状态
	infoCmd := exec.Command("docker", "info")
	infoCmd.Env = append(os.Environ(), "DOCKER_CLI_HINTS=false")
	if err := infoCmd.Run(); err != nil {
		return true, version, fmt.Errorf(t(i18n.CliCheckErrDaemonNotRunning), "docker")
	}

	return true, version, nil
}

// CheckDockerCompose 检查 docker compose 是否安装并获取版本信息。
// 优先尝试使用官方 V2 插件 `docker compose`，如果失败则回退寻找旧的 `docker-compose` 独立二级制文件。
func CheckDockerCompose() (bool, string, error) {
	t := clii18n.T

	var cmd *exec.Cmd
	var toolName string

	// 先尝试 v2 环境 (docker compose)
	if _, err := exec.LookPath("docker"); err == nil {
		cmd = exec.Command("docker", "compose", "version")
		toolName = "docker compose"

		// 有时候 docker 存在，但插件 compose 不存在。我们测试跑一下能否运行
		if err := cmd.Run(); err != nil {
			cmd = nil // V2 不可用
		}
	}

	// 若 v2 指令未能成功或不存在，降级尝试独立的 v1 二进制程序 (docker-compose)
	if cmd == nil {
		if _, err := exec.LookPath("docker-compose"); err == nil {
			cmd = exec.Command("docker-compose", "--version")
			toolName = "docker-compose"
		} else {
			return false, "", fmt.Errorf(t(i18n.CliCheckErrNotFound), "docker-compose")
		}
	}

	// 正式执行版本获取动作
	finalCmd := exec.Command(cmd.Args[0], cmd.Args[1:]...)
	output, err := finalCmd.CombinedOutput()
	if err != nil {
		return true, "", fmt.Errorf(t(i18n.CliCheckErrExecFailed), toolName, err)
	}

	outStr := strings.TrimSpace(string(output))

	// 输出可能是:
	// v2: "Docker Compose version v2.21.0-desktop.1"
	// v1: "docker-compose version 1.29.2, build 5becea4c"
	re := regexp.MustCompile(`[Vv]ersion\s+([vV]?\d+\.\d+\.\d+(?:[-a-zA-Z0-9.]+)*)`)
	matches := re.FindStringSubmatch(outStr)

	if len(matches) > 1 {
		return true, matches[1], nil
	}

	return true, outStr, nil
}

// CheckDB 验证数据库连通性
func CheckDB(dsn string) error {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrDbConn), err)
	}
	defer sqlDB.Close()

	sqlDB.SetConnMaxLifetime(3 * time.Second)
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrDbConn), err)
	}
	return nil
}

// CheckZentao 验证禅道服务有效性
func CheckZentao(apiURL string) error {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrZentaoConn), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrZentaoConn), fmt.Sprintf("HTTP %d", resp.StatusCode))
	}

	return nil
}

// PortStatus 描述单个端口的占用情况
type PortStatus struct {
	Port    string
	InUse   bool
	PID     string
	Process string
	Unknown bool // 检测失败时为 true
}

// CheckPort 检查指定端口是否被占用，并返回占用者的 PID 和进程名。
// 跨平台支持：Linux 使用 ss/lsof，Windows 使用 netstat。
func CheckPort(port string) PortStatus {
	result := PortStatus{Port: port}

	switch runtime.GOOS {
	case "linux":
		result = checkPortLinux(port)
	case "windows":
		result = checkPortWindows(port) // → windows.go
	default:
		result.Unknown = true
	}
	return result
}

func checkPortLinux(port string) PortStatus {
	result := PortStatus{Port: port}
	// 优先使用 ss（比 netstat 更现代）
	out, err := exec.Command("ss", "-tlnp", "sport", "=", ":"+port).Output()
	if err != nil {
		// 降级使用 lsof
		out, err = exec.Command("lsof", "-i", ":"+port, "-sTCP:LISTEN", "-n", "-P").Output()
		if err != nil {
			result.Unknown = true
			return result
		}
		return parseLsofOutput(string(out), port)
	}
	return parseSsOutput(string(out), port)
}

// parseLsofOutput 解析 lsof 输出提取 PID / 进程名
func parseLsofOutput(output, port string) PortStatus {
	result := PortStatus{Port: port}
	for i, line := range strings.Split(output, "\n") {
		if i == 0 {
			continue // 跳过 header
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		result.InUse = true
		result.Process = fields[0]
		result.PID = fields[1]
		return result
	}
	return result
}

// parseSsOutput 解析 ss 输出提取 PID / 进程名
// ss -tlnp 输出格式: State Recv-Q Send-Q Local ... Process
func parseSsOutput(output, port string) PortStatus {
	result := PortStatus{Port: port}
	pidRe := regexp.MustCompile(`pid=(\d+)`)
	nameRe := regexp.MustCompile(`"([^"]+)"`)
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, ":"+port) {
			continue
		}
		result.InUse = true
		if m := pidRe.FindStringSubmatch(line); len(m) > 1 {
			result.PID = m[1]
		}
		if m := nameRe.FindStringSubmatch(line); len(m) > 1 {
			result.Process = m[1]
		}
		return result
	}
	return result
}
