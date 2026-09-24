// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package common

import (
	"github.com/bmatcuk/doublestar"
)

// CheckFilesMatchPatterns 检查是否有文件匹配任一模式
func CheckFilesMatchPatterns(files []string, patterns []string) bool {
	// 如果没有指定文件或者模式，则默认匹配
	if len(files) == 0 || len(patterns) == 0 {
		return true
	}

	// 检查是否有任一文件匹配任一模式
	for _, file := range files {
		for _, pattern := range patterns {
			matched, err := doublestar.Match(pattern, file)

			if err == nil && matched {
				return true
			}
		}
	}
	// 没有文件匹配任何模式
	return false
}
