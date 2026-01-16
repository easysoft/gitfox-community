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
