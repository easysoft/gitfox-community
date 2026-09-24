// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package git

type Repository interface {
	GetGitUID() string
}

// CreateReadParams creates base read parameters for git read operations.
// IMPORTANT: repo is assumed to be not nil!
// CreateReadParams为git的读操作创建基本读参数。
// 重要：假定repo不是nil！
func CreateReadParams(repo Repository) ReadParams {
	return ReadParams{
		RepoUID: repo.GetGitUID(),
	}
}
