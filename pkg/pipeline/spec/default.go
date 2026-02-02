package yaml

type Default struct {
	Clone      *Clone            `json:"clone,omitempty"`
	Repository *Repository       `json:"repository,omitempty"`
	Registry   *Registry         `json:"registry,omitempty"`
	Resources  *Resources        `json:"resources,omitempty"`
	Delegate   Stringorslice     `json:"delegate,omitempty"`
	Barriers   []*Barrier        `json:"barriers,omitempty"`
	Status     *Status           `json:"status,omitempty"`
	Timeout    string            `json:"timeout,omitempty"`
	Envs       map[string]string `json:"envs,omitempty"`
}
