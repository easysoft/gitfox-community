// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package utils

import (
	neturl "net/url"
)

// DecodeGitRef 对可能被 URL 编码的 Git 引用（分支名/标签名）进行解码。
// 前端传递含 '/' 的分支名（如 mg/test01）时可能被编码为 mg%2Ftest01，
// 直接传给 git 命令会导致 'ambiguous argument' 错误。
func DecodeGitRef(ref string) string {
	if ref == "" {
		return ref
	}
	if decoded, err := neturl.PathUnescape(ref); err == nil {
		return decoded
	}
	return ref
}
