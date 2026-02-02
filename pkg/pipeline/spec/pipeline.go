package yaml

// Pipeline defines the pipeline execution.
type Pipeline struct {
	Stages  []*Stage          `json:"stages,omitempty"`
	Inputs  map[string]*Input `json:"inputs,omitempty"`
	Options *Default          `json:"options,omitempty"`
}
