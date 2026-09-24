// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package rbac

const (
	MatchWildcard          = "*"
	MatchRecursiveWildcard = "**"
)

type Effect string

const (
	EffectDeny  Effect = "deny"
	EffectAllow Effect = "allow"
)

type WildcardLevel int

const (
	WildcardLevelRoot WildcardLevel = iota
	WildcardLevelSub
	WildcardLevelRecursive
)

type PolicyItemIndex int

const (
	PolicyItemSub PolicyItemIndex = iota
	PolicyItemSpace
	PolicyItemObj
	PolicyItemAct
	PolicyItemEft
	PolicyItemPriority
)
