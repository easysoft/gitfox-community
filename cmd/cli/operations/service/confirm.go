// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// confirmPrompt 弹出一个 [Y/N] 确认提示，不设置默认值，用户必须显式输入 Y 或 N。
// 返回 (true, nil) 表示用户确认，(false, nil) 表示用户拒绝。
// 当用户按下 Ctrl+C 或输入 EOF 时，返回 (false, huh.ErrUserAborted)。
func confirmPrompt(title string) (bool, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stdout, "%s [Y/N] ", title)
		if !scanner.Scan() {
			fmt.Fprintln(os.Stdout)
			if err := scanner.Err(); err != nil {
				return false, err
			}
			return false, huh.ErrUserAborted
		}
		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(os.Stdout, t(i18n.CliInitConfirmInvalidInput))
		}
	}
}

// inputPrompt 弹出一个单行输入提示，支持默认值和校验。
// 如果 defaultValue 非空，在提示中显示并在用户直接回车时使用默认值。
// validate 为 nil 时不做校验。
func inputPrompt(title string, defaultValue string, validate func(string) error) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if defaultValue != "" {
			fmt.Fprintf(os.Stdout, "%s [%s]: ", title, defaultValue)
		} else {
			fmt.Fprintf(os.Stdout, "%s: ", title)
		}
		if !scanner.Scan() {
			fmt.Fprintln(os.Stdout)
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", huh.ErrUserAborted
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" && defaultValue != "" {
			input = defaultValue
		}
		if validate != nil {
			if err := validate(input); err != nil {
				fmt.Fprintln(os.Stdout, err.Error())
				continue
			}
		}
		return input, nil
	}
}

// passwordPrompt 弹出一个密码输入提示，输入内容不回显。
func passwordPrompt(title string, validate func(string) error) (string, error) {
	for {
		fmt.Fprintf(os.Stdout, "%s: ", title)
		pwdBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stdout) // 换行（ReadPassword 不会自动换行）
		if err != nil {
			// 区分终端不可用等真实 IO 错误与用户中断
			if !term.IsTerminal(int(syscall.Stdin)) {
				return "", fmt.Errorf("stdin is not a terminal: %w", err)
			}
			return "", huh.ErrUserAborted
		}
		input := string(pwdBytes)
		if validate != nil {
			if err := validate(input); err != nil {
				fmt.Fprintln(os.Stdout, err.Error())
				continue
			}
		}
		return input, nil
	}
}
