package pipeline

type CiCacheMetadata struct {
	Key         string `json:"key"`
	Size        int64  `json:"size"`
	ExecutionID int64  `json:"execution_id"`
	Repo        string `json:"repo"`
	RepoId      int64  `json:"repo_id"`
	Space       string `json:"space"`
	SpaceId     int64  `json:"space_id"`
	CreatedAt   int64  `json:"create_at"`
	PulledAt    int64  `json:"pulled_at"`
	Downloads   int    `json:"downloads"`
}
