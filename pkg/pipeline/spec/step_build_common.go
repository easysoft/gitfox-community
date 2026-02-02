package yaml

type StepBuildCommon struct {
	Script    Stringorslice     `json:"script,omitempty"`
	Container *Container        `json:"container,omitempty"`
	Envs      map[string]string `json:"envs,omitempty"`
	Mount     []*Mount          `json:"mount,omitempty"`
	Caching   *Caching          `json:"caching,omitempty"`
	Mirror    Stringorslice     `json:"mirror,omitempty"`
	Proxy     *Proxy            `json:"proxy,omitempty"`
	Version   string            `json:"version,omitempty"`
	BuildEnv  string            `json:"buildEnv,omitempty"`
}
