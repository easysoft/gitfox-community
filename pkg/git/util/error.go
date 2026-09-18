// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package util

import (
	"context"
	stderr "errors"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/git/command"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// ProcessGitErrorf 处理 Git 错误并转换为应用层 GitError 类型错误
func ProcessGitErrorf(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// 获取错误信息用于匹配
	errMsg := err.Error()

	// 记录原始错误
	logger.GetLogger(ctx).Warnw("git error occurred", "error", err)

	// 识别特定的 Git 错误类型，并包装底层错误
	// 按照错误出现频率排序，最常见的错误放在前面以提高性能
	var translatedError error
	switch {
	// 1.1. 最常见：引用/分支/标签不存在
	case isRefNotFoundError(errMsg):
		translatedError = errors.ErrGitRefNotFound
	// 1.2. 最常见：分支不存在
	case isBranchRefNotFoundError(errMsg):
		translatedError = errors.ErrGitBranchNotFound
	// 1.3. 最常见：标签不存在
	case isTagRefNotFoundError(errMsg):
		translatedError = errors.ErrGitTagNotFound

	// 2.0. 分支 ref 路径冲突（如已有 test 分支时创建 test/haha）
	case isBranchRefConflictError(errMsg):
		translatedError = errors.ErrGitBranchRefConflict
	// 2.1. 常见：引用已存在
	case isRefAlreadyExistsError(errMsg):
		translatedError = errors.ErrGitRefAlreadyExists
	// 2.2. 常见：分支已存在
	case isRefBranchAlreadyExistsError(errMsg):
		translatedError = errors.ErrGitBranchRefAlreadyExists
	// 2.3. 常见：标签已存在
	case isRefTagAlreadyExistsError(errMsg):
		translatedError = errors.ErrGitTagRefAlreadyExists
	// 3. 常见：仓库不存在
	case isRepoNotFoundError(errMsg):
		translatedError = errors.ErrGitRepoNotFound

	// 4. 常见：无效的引用
	case isInvalidRefError(err, errMsg):
		translatedError = errors.ErrGitInvalidRef

	// 5. 常见：引用不明确
	case isAmbiguousRefError(err, errMsg):
		translatedError = errors.ErrGitAmbiguousRef

	// 6. 空仓库
	case isEmptyRepoError(errMsg):
		translatedError = errors.ErrGitEmptyRepo

	// 7. 推送被拒绝
	case isPushRejectedError(errMsg):
		translatedError = errors.ErrGitPushRejected

	// 8. 推送过期
	case isPushOutOfDateError(errMsg):
		translatedError = errors.ErrGitPushOutOfDate

	// 9. 合并冲突
	case isMergeConflictError(errMsg):
		translatedError = errors.ErrGitMergeConflict

	// 10. 不相关的历史
	case isUnrelatedHistoriesError(errMsg):
		translatedError = errors.ErrGitUnrelatedHistories

	// 11. 路径不存在
	case isPathNotFoundError(errMsg):
		translatedError = errors.ErrGitPathNotFound

	// 12. 无效的路径
	case isInvalidPathError(errMsg):
		translatedError = errors.ErrGitInvalidPath

	// 13. 对象不存在
	case isObjectNotFoundError(errMsg):
		translatedError = errors.ErrGitObjectNotFound

	// 14. 无效的签名
	case isInvalidSignatureError(errMsg):
		translatedError = errors.ErrGitInvalidSignature

	// 15. Git 版本不支持该参数（如 --atomic 等新版本才有的选项）
	case isUnknownOptionError(errMsg):
		optionName := extractUnknownOption(errMsg)
		if optionName != "" {
			translatedError = errors.GitError.Newf(
				"current git version does not support option '%s', please upgrade git",
				optionName,
			)
		} else {
			translatedError = errors.ErrGitUnsupportedOption
		}

	// 16. 默认：其他 Git 错误
	default:
		translatedError = errors.ErrGitCommandFailed
	}

	return translatedError
}

// isRefNotFoundError 检查是否是引用不存在错误
func isRefNotFoundError(errMsg string) bool {
	// 检查常见的引用不存在错误消息
	if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "ref") {
		return true
	}

	if strings.Contains(errMsg, "unknown revision") {
		return true
	}

	if strings.Contains(errMsg, "couldn't find remote ref") {
		return true
	}
	return false
}

// isBranchRefNotFoundError 检查是否是分支引用不存在错误
func isBranchRefNotFoundError(errMsg string) bool {
	// 检查常见的引用不存在错误消息
	if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "branch") {
		return true

	}
	return false
}

// isTagRefNotFoundError 检查是否是标签引用不存在错误
func isTagRefNotFoundError(errMsg string) bool {
	// 检查常见的引用不存在错误消息
	if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "tag") {
		return true
	}

	return false
}

// isRefAlreadyExistsError 检查是否是引用已存在错误
func isRefAlreadyExistsError(errMsg string) bool {
	errMsgLower := strings.ToLower(errMsg)
	// 匹配 "reference already exists"
	if strings.Contains(errMsgLower, "reference already exists") {
		return true
	}
	// 匹配 ""ref 'xxx' already exists"
	if strings.Contains(errMsgLower, "already exists") && strings.Contains(errMsgLower, "ref") {
		return true
	}
	return false
}

// isBranchRefConflictError 检查是否是分支 ref 路径冲突错误
// 当已有分支 test 时创建 test/haha，或已有 test/haha 时创建 test，git 会报此错误
// 原因：git ref 存储中文件和目录不能同名共存
func isBranchRefConflictError(errMsg string) bool {
	errMsgLower := strings.ToLower(errMsg)
	return strings.Contains(errMsgLower, "cannot lock ref") && strings.Contains(errMsgLower, "exists; cannot create")
}

// isRefBranchAlreadyExistsError 检查是否是分支引用已存在错误
func isRefBranchAlreadyExistsError(errMsg string) bool {
	errMsgLower := strings.ToLower(errMsg)
	// 匹配  "branch 'xxx' already exists"
	if strings.Contains(errMsgLower, "already exists") && strings.Contains(errMsgLower, "branch") {
		return true
	}
	return false
}

// isRefTagAlreadyExistsError 检查是否是标签引用已存在错误
func isRefTagAlreadyExistsError(errMsg string) bool {
	errMsgLower := strings.ToLower(errMsg)

	// 匹配 "tag 'xxx' already exists"
	if strings.Contains(errMsgLower, "already exists") && strings.Contains(errMsgLower, "tag") {
		return true
	}
	return false
}

// isRepoNotFoundError 检查是否是仓库不存在错误
func isRepoNotFoundError(errMsg string) bool {
	return strings.Contains(errMsg, "repository not found") ||
		strings.Contains(errMsg, "no such file or directory") ||
		strings.Contains(errMsg, "does not exist") && strings.Contains(errMsg, "repository")
}

// isInvalidRefError 检查是否是无效引用错误
func isInvalidRefError(err error, errMsg string) bool {
	// 检查 command.Error 类型的无效引用错误
	var cmdErr *command.Error
	if stderr.As(err, &cmdErr) && cmdErr.IsInvalidRefErr() {
		return true
	}
	return strings.Contains(errMsg, "not a valid ref") ||
		strings.Contains(errMsg, "invalid ref")
}

// isAmbiguousRefError 检查是否是引用不明确错误
func isAmbiguousRefError(err error, errMsg string) bool {
	// 检查 command.Error 类型的不明确引用错误
	var cmdErr *command.Error
	if stderr.As(err, &cmdErr) && cmdErr.IsAmbiguousArgErr() {
		return true
	}
	return strings.Contains(errMsg, "ambiguous argument") ||
		strings.Contains(errMsg, "ambiguous ref")
}

// isEmptyRepoError 检查是否是空仓库错误
func isEmptyRepoError(errMsg string) bool {
	return strings.Contains(errMsg, "empty repository") ||
		strings.Contains(errMsg, "does not have any commits")
}

// isPushRejectedError 检查是否是推送被拒绝错误
func isPushRejectedError(errMsg string) bool {
	return strings.Contains(errMsg, "push rejected") ||
		strings.Contains(errMsg, "[rejected]") ||
		strings.Contains(errMsg, "pre-receive hook declined")
}

// isPushOutOfDateError 检查是否是推送过期错误
func isPushOutOfDateError(errMsg string) bool {
	return strings.Contains(errMsg, "non-fast-forward") ||
		strings.Contains(errMsg, "fetch first") ||
		strings.Contains(errMsg, "Updates were rejected")
}

// isMergeConflictError 检查是否是合并冲突错误
func isMergeConflictError(errMsg string) bool {
	return strings.Contains(errMsg, "merge conflict") ||
		strings.Contains(errMsg, "CONFLICT") ||
		strings.Contains(errMsg, "Automatic merge failed")
}

// isUnrelatedHistoriesError 检查是否是不相关历史错误
func isUnrelatedHistoriesError(errMsg string) bool {
	return strings.Contains(errMsg, "no merge base") ||
		strings.Contains(errMsg, "unrelated histories")
}

// isPathNotFoundError 检查是否是路径不存在错误
func isPathNotFoundError(errMsg string) bool {
	return strings.Contains(errMsg, "path") && strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "path") && strings.Contains(errMsg, "does not exist")
}

// isInvalidPathError 检查是否是无效路径错误
func isInvalidPathError(errMsg string) bool {
	return strings.Contains(errMsg, "path is invalid") ||
		strings.Contains(errMsg, "invalid path")
}

// isObjectNotFoundError 检查是否是对象不存在错误
func isObjectNotFoundError(errMsg string) bool {
	return strings.Contains(errMsg, "object not found") ||
		strings.Contains(errMsg, "missing object") ||
		strings.Contains(errMsg, "bad object")
}

// isInvalidSignatureError 检查是否是无效签名错误
func isInvalidSignatureError(errMsg string) bool {
	return strings.Contains(errMsg, "invalid signature") ||
		strings.Contains(errMsg, "bad signature")
}

// isUnknownOptionError 检查是否是 Git 版本不支持某参数的错误
// 例如: unknown option `atomic'
func isUnknownOptionError(errMsg string) bool {
	return strings.Contains(errMsg, "unknown option")
}

// extractUnknownOption 从 Git 错误信息中提取不支持的参数名
// 例如从 "unknown option `atomic'" 中提取 "atomic"
func extractUnknownOption(errMsg string) string {
	// 尝试匹配 backtick 包裹的选项名: `option'
	const prefix = "unknown option `"
	start := strings.Index(errMsg, prefix)
	if start == -1 {
		// 尝试匹配单引号包裹的选项名: 'option'
		const altPrefix = "unknown option '"
		start = strings.Index(errMsg, altPrefix)
		if start == -1 {
			return ""
		}
		start += len(altPrefix)
		end := strings.Index(errMsg[start:], "'")
		if end == -1 {
			return ""
		}
		return errMsg[start : start+end]
	}
	start += len(prefix)
	end := strings.Index(errMsg[start:], "'")
	if end == -1 {
		return ""
	}
	return errMsg[start : start+end]
}
