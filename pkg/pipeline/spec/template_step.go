package yaml

// TemplateStep defines a step-level template
type TemplateStep struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Step        *Step             `json:"step,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}
