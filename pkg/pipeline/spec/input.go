package yaml

// Input defines an input parameter.
type Input struct {
	Type        string      `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Items       *InputItems `json:"items,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Mask        bool        `json:"mask,omitempty"`
}
