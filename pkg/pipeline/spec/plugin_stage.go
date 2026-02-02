package yaml

// PluginStage defines a stage-level plugin
type PluginStage struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Stage       *Stage            `json:"stage,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}
