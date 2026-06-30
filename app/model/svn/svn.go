package svn

// FuncType 是受白名单约束的 SVN 子命令名。
type FuncType string

const (
	FuncLs    FuncType = "ls"
	FuncLog   FuncType = "log"
	FuncBlame FuncType = "blame"
	FuncCat   FuncType = "cat"
	FuncInfo  FuncType = "info"
	FuncDiff  FuncType = "diff"
)

// IsValid 判断 FuncType 是否在白名单内。
func (f FuncType) IsValid() bool {
	switch f {
	case FuncLs, FuncLog, FuncBlame, FuncCat, FuncInfo, FuncDiff:
		return true
	}
	return false
}

// ReqForExec 是 POST /api/v2/repos/:id/svn/exec 的统一请求体。
type ReqForExec struct {
	Func FuncType       `json:"func" binding:"required"`
	Args map[string]any `json:"args"`
}
