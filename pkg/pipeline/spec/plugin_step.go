package yaml

// PluginStep defines a step-level plugin
type PluginStep struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Step        *Step             `json:"step,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}
