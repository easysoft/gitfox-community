package events

const (
	// category defines the event category used for this package.
	category = "pullreq"
)

type Base struct {
	PullReqID    int64  `json:"pullreqID"`
	SourceRepoID int64  `json:"sourceRepoID"`
	TargetRepoID int64  `json:"targetRepoID"`
	Account      string `json:"account"`
}
