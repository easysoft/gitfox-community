package yaml

// TemplateStage defines a stage-level template
type TemplateStage struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Stage       *Stage            `json:"stage,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}
