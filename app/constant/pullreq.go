// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

import (
	gitenum "github.com/easysoft/gitfox-community/pkg/git/enum"
)

// PullReqState defines pull request state.
type PullReqState string

func (PullReqState) Enum() []interface{}                  { return toInterfaceSlice(pullReqStates) }
func (s PullReqState) Sanitize() (PullReqState, bool)     { return Sanitize(s, GetAllPullReqStates) }
func GetAllPullReqStates() ([]PullReqState, PullReqState) { return pullReqStates, "" }

// PullReqState enumeration.
const (
	PullReqStateOpen   PullReqState = "open"
	PullReqStateMerged PullReqState = "merged"
	PullReqStateClosed PullReqState = "closed"
)

var pullReqStates = sortEnum([]PullReqState{
	PullReqStateOpen,
	PullReqStateMerged,
	PullReqStateClosed,
})

type MergeMethod gitenum.MergeMethod

// MergeMethod enumeration.
const (
	MergeMethodMerge       = MergeMethod(gitenum.MergeMethodMerge)
	MergeMethodSquash      = MergeMethod(gitenum.MergeMethodSquash)
	MergeMethodRebase      = MergeMethod(gitenum.MergeMethodRebase)
	MergeMethodFastForward = MergeMethod(gitenum.MergeMethodFastForward)
)

var MergeMethods = sortEnum([]MergeMethod{
	MergeMethodMerge,
	MergeMethodSquash,
	MergeMethodRebase,
	MergeMethodFastForward,
})

func (MergeMethod) Enum() []interface{} { return toInterfaceSlice(MergeMethods) }
func (m MergeMethod) Sanitize() (MergeMethod, bool) {
	s, ok := gitenum.MergeMethod(m).Sanitize()
	return MergeMethod(s), ok
}
