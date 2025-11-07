package settings

type Setting struct {
	SpaceID int    `json:"space_id"`
	RepoID  int    `json:"repo_id"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

type SystemSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
