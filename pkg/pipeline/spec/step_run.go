package yaml

type StepRun struct {
	Shell     string            `json:"shell,omitempty"`
	Script    Stringorslice     `json:"script,omitempty"`
	Container *Container        `json:"container,omitempty"`
	Envs      map[string]string `json:"envs,omitempty"`
	Reports   []*Report         `json:"reports,omitempty"`
	Outputs   []string          `json:"outputs,omitempty"`
	Mount     []*Mount          `json:"mount,omitempty"`
}
