package yaml

type StepScript struct {
	Image      string            `json:"image,omitempty"`
	Connector  string            `json:"connector,omitempty"`
	User       string            `json:"user,omitempty"`
	Group      string            `json:"group,omitempty"`
	Pull       string            `json:"pull,omitempty"`
	Shell      string            `json:"shell,omitempty"`
	Envs       map[string]string `json:"envs,omitempty"`
	Run        string            `json:"run,omitempty"`
	Entry      string            `json:"entry,omitempty"`
	Entrypoint string            `json:"entrypoint,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Privileged bool              `json:"privileged,omitempty"`
	Network    string            `json:"network,omitempty"`
	Reports    []*Report         `json:"reports,omitempty"`
	Outputs    []string          `json:"outputs,omitempty"`
	Resources  *Resources        `json:"resources,omitempty"`
	Mount      []*Mount          `json:"mount,omitempty"`
}
