package yaml

type StepExec struct {
	Shell  string            `json:"shell,omitempty"`
	Script Stringorslice     `json:"script,omitempty"`
	Envs   map[string]string `json:"envs,omitempty"`
}
