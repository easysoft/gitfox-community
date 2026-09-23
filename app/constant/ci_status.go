// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

import "strings"

// CIStatus defines the different kinds of CI statuses for
// stages, steps and executions.
type CIStatus string

const (
	CIStatusSkipped       CIStatus = "skipped"
	CIStatusWaitingOnDeps CIStatus = "waiting_on_dependencies"
	CIStatusPending       CIStatus = "pending"
	CIStatusRunning       CIStatus = "running"
	CIStatusSuccess       CIStatus = "success"
	CIStatusFailure       CIStatus = "failure"
	CIStatusKilled        CIStatus = "killed"
	CIStatusError         CIStatus = "error"
	CIStatusBlocked       CIStatus = "blocked"
	CIStatusDeclined      CIStatus = "declined"
)

// Enum returns all possible CIStatus values.
func (CIStatus) Enum() []interface{} {
	return toInterfaceSlice(ciStatuses)
}

// Sanitize validates and returns a sanitized CIStatus value.
func (status CIStatus) Sanitize() (CIStatus, bool) {
	return Sanitize(status, GetAllCIStatuses)
}

// GetAllCIStatuses returns all possible CIStatus values and a default value.
func GetAllCIStatuses() ([]CIStatus, CIStatus) {
	return ciStatuses, CIStatusPending
}

func (status CIStatus) ConvertToCheckStatus() CheckStatus {
	if status == CIStatusPending || status == CIStatusWaitingOnDeps {
		return CheckStatusPending
	}
	if status == CIStatusSuccess || status == CIStatusSkipped {
		return CheckStatusSuccess
	}
	if status == CIStatusFailure {
		return CheckStatusFailure
	}
	if status == CIStatusRunning {
		return CheckStatusRunning
	}
	return CheckStatusError
}

// ParseCIStatus converts the status from a string to typed enum.
// If the match is not exact, will just return default error status
// instead of explicitly returning not found error.
func ParseCIStatus(status string) CIStatus {
	switch strings.ToLower(status) {
	case "skipped", "blocked", "declined", "waiting_on_dependencies",
		"pending", "running", "success", "failure", "killed", "error":
		return CIStatus(strings.ToLower(status))
	case "": // just in case status is not passed through
		return CIStatusPending
	default:
		return CIStatusError
	}
}

// IsDone returns true if in a completed state.
func (status CIStatus) IsDone() bool {
	//nolint:exhaustive
	switch status {
	case CIStatusWaitingOnDeps,
		CIStatusPending,
		CIStatusRunning,
		CIStatusBlocked:
		return false
	default:
		return true
	}
}

// IsFailed returns true if in a failed state.
func (status CIStatus) IsFailed() bool {
	return status == CIStatusFailure ||
		status == CIStatusKilled ||
		status == CIStatusError
}

// IsCancelable returns true if in a pending or running state.
func (status CIStatus) IsCancelable() bool {
	return status == CIStatusPending || status == CIStatusRunning
}

// List of all CIStatus values.
var ciStatuses = sortEnum([]CIStatus{
	CIStatusSkipped,
	CIStatusBlocked,
	CIStatusDeclined,
	CIStatusWaitingOnDeps,
	CIStatusPending,
	CIStatusRunning,
	CIStatusSuccess,
	CIStatusFailure,
	CIStatusKilled,
	CIStatusError,
})
