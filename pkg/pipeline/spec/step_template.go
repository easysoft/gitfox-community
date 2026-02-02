package yaml

type StepTemplate struct {
	Name     string                 `json:"name,omitempty"`
	Inputs   map[string]interface{} `json:"inputs,omitempty"`
	Overlays map[string]interface{} `json:"overlays,omitempty"`
}
