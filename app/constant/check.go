// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

import "golang.org/x/exp/slices"

// CheckStatus defines status check status.
type CheckStatus string

func (CheckStatus) Enum() []interface{}                 { return toInterfaceSlice(checkStatuses) }
func (s CheckStatus) Sanitize() (CheckStatus, bool)     { return Sanitize(s, GetAllCheckStatuses) }
func GetAllCheckStatuses() ([]CheckStatus, CheckStatus) { return checkStatuses, "" }

// CheckStatus enumeration.
const (
	CheckStatusPending CheckStatus = "pending"
	CheckStatusRunning CheckStatus = "running"
	CheckStatusSuccess CheckStatus = "success"
	CheckStatusFailure CheckStatus = "failure"
	CheckStatusError   CheckStatus = "error"
)

var checkStatuses = sortEnum([]CheckStatus{
	CheckStatusPending,
	CheckStatusRunning,
	CheckStatusSuccess,
	CheckStatusFailure,
	CheckStatusError,
})

var terminalCheckStatuses = []CheckStatus{CheckStatusFailure, CheckStatusSuccess, CheckStatusError}

// CheckPayloadKind defines status payload type.
type CheckPayloadKind string

func (CheckPayloadKind) Enum() []interface{} { return toInterfaceSlice(checkPayloadTypes) }
func (s CheckPayloadKind) Sanitize() (CheckPayloadKind, bool) {
	return Sanitize(s, GetAllCheckPayloadTypes)
}
func GetAllCheckPayloadTypes() ([]CheckPayloadKind, CheckPayloadKind) {
	return checkPayloadTypes, CheckPayloadKindEmpty
}

// CheckPayloadKind enumeration.
const (
	CheckPayloadKindEmpty    CheckPayloadKind = ""
	CheckPayloadKindRaw      CheckPayloadKind = "raw"
	CheckPayloadKindMarkdown CheckPayloadKind = "markdown"
	CheckPayloadKindPipeline CheckPayloadKind = "pipeline"
)

var checkPayloadTypes = sortEnum([]CheckPayloadKind{
	CheckPayloadKindEmpty,
	CheckPayloadKindRaw,
	CheckPayloadKindMarkdown,
	CheckPayloadKindPipeline,
})

func (s CheckStatus) IsCompleted() bool {
	return slices.Contains(terminalCheckStatuses, s)
}
