// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package event

// SSEType defines the kind of server sent event.
type SSEType string

// Enums for event types delivered to the event stream for the UI.
const (
	SSETypeExecutionUpdated   SSEType = "execution_updated"
	SSETypeExecutionRunning   SSEType = "execution_running"
	SSETypeExecutionCompleted SSEType = "execution_completed"
	SSETypeExecutionCanceled  SSEType = "execution_canceled"

	SSETypeRepositoryImportCompleted SSEType = "repository_import_completed"
	SSETypeRepositoryExportCompleted SSEType = "repository_export_completed"

	SSETypePullRequestUpdated         SSEType = "pullreq_updated"
	SSETypePullRequestReviewerAdded   SSEType = "pullreq_reviewer_added"
	SSETypePullRequestReviewerRemoved SSEType = "pullreq_reviewer_removed"

	SSETypeLogLineAppended SSEType = "log_line_appended"
)
