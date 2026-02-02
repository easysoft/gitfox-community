package yaml

type StepBuild struct {
	Shell       string            `json:"shell,omitempty"`
	Script      Stringorslice     `json:"script,omitempty"`
	Container   *Container        `json:"container,omitempty"`
	Envs        map[string]string `json:"envs,omitempty"`
	Reports     []*Report         `json:"reports,omitempty"`
	Outputs     []string          `json:"outputs,omitempty"`
	Mount       []*Mount          `json:"mount,omitempty"`
	Caching     *Caching          `json:"caching,omitempty"`
	PreActions  []string          `json:"preActions,omitempty"`
	PostActions []string          `json:"postActions,omitempty"`
	BuildEnv    string            `json:"buildEnv,omitempty"`
}
