package yaml

type StatusStage struct {
	Disabled bool   `json:"disabled,omitempty"`
	Matrix   string `json:"matrix,omitempty"`
	Name     string `json:"name,omitempty"`
}
