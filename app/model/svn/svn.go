// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package svn

// Command 是受白名单约束的 SVN 子命令名。
type Command string

const (
	Ls    Command = "ls"
	Log   Command = "log"
	Blame Command = "blame"
	Cat   Command = "cat"
	Info  Command = "info"
	Diff  Command = "diff"
)

// IsValid 判断 Command 是否在白名单内。
func (f Command) IsValid() bool {
	switch f {
	case Ls, Log, Blame, Cat, Info, Diff:
		return true
	}
	return false
}

// ReqForExec 是 POST /api/v2/repos/:id/svn/exec 的统一请求体。
type ReqForExec struct {
	Command Command        `json:"command" binding:"required"`
	Args    map[string]any `json:"args"`
}
