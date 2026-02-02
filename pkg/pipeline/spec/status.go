package yaml

type Status struct {
	Disabled bool   `json:"disabled,omitempty"`
	Level    string `json:"level,omitempty"`
	Matrix   string `json:"matrix,omitempty"`
	Name     string `json:"name,omitempty"`
}
