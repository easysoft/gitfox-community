// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package service

// windows.go 汇总了 GitFox CLI 在 Windows 平台独有的通用辅助逻辑。
//
// 本文件中的所有函数不依赖 golang.org/x/sys/windows/svc，因此无需构建标签隔离，
// 可在任何平台编译（但仅在 Windows 运行时有实际意义）。
//
// 平台独有的服务管理核心逻辑（注册/停止/卸载/运行）已迁移至：
//   - windows_svc.go      (//go:build windows)
//   - windows_svc_stub.go (//go:build !windows)
//
// 本文件保留以下通用逻辑：
//   - 端口检测：checkPortWindows
//   - 日志/状态命令提示：windowsLogCmd / windowsStatusCmd
//   - 权限校验：checkSudoWindows
//   - 平台常量：windowsBinaryName / windowsCmdPrefix

import (
	"os/exec"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/i18n"
	clii18n "github.com/easysoft/gitfox-community/pkg/i18n/cli"
)

// ─────────────────────────────────────────────
// Windows 平台常量
// ─────────────────────────────────────────────

// windowsBinaryName 是 Windows 平台下 GitFox 可执行文件名
const windowsBinaryName = "gitfox.exe"

// windowsCmdPrefix 是 Windows 平台下在 Shell 中调用 GitFox 的前缀写法
const windowsCmdPrefix = "gitfox.exe"

// checkSudoWindows 的实现位于 windows_svc.go（使用 Windows API 检测管理员权限），
// 非 Windows 平台的桩函数位于 windows_svc_stub.go。

// ─────────────────────────────────────────────
// 端口检测
// ─────────────────────────────────────────────

// checkPortWindows 使用 netstat + tasklist 检测指定端口是否被占用，并返回占用进程信息。
func checkPortWindows(port string) PortStatus {
	result := PortStatus{Port: port}

	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		result.Unknown = true
		return result
	}

	for _, line := range strings.Split(string(out), "\n") {
		// 只关心 LISTENING 状态的目标端口行
		if !strings.Contains(line, ":"+port) || !strings.Contains(line, "LISTENING") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		result.InUse = true
		result.PID = fields[4]

		// 通过 tasklist 查进程名
		if nameOut, err := exec.Command(
			"tasklist", "/FI", "PID eq "+result.PID, "/NH", "/FO", "CSV",
		).Output(); err == nil {
			parts := strings.Split(strings.TrimSpace(string(nameOut)), ",")
			if len(parts) > 0 {
				result.Process = strings.Trim(parts[0], "\"")
			}
		}

		return result
	}

	return result
}

// ─────────────────────────────────────────────
// 日志 / 状态命令提示
// ─────────────────────────────────────────────

// windowsLogCmd 返回 Windows 平台下查看 GitFox 服务日志的推荐命令字符串。
func windowsLogCmd() string {
	return clii18n.T(i18n.CliStartSummaryLogCmdWindows)
}

// windowsStatusCmd 返回 Windows 平台下查询 GitFox 服务运行状态的推荐命令字符串。
func windowsStatusCmd() string {
	return "sc query gitfox"
}
