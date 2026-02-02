package yaml

type StageCI struct {
	Cache    *Cache            `json:"cache,omitempty"`
	Clone    *CloneStage       `json:"clone,omitempty"`
	Platform *Platform         `json:"platform,omitempty"`
	Runtime  *Runtime          `json:"runtime,omitempty"`
	Steps    []*Step           `json:"steps,omitempty"`
	Envs     map[string]string `json:"envs,omitempty"`
	Volumes  []*Volume         `json:"volumes,omitempty"`
	Vars     []*Variable       `json:"vars,omitempty"`
}
