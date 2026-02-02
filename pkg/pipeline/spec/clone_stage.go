package yaml

// Clone defines the stage clone behavior.
type CloneStage struct {
	Depth    int64  `json:"depth,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
	Strategy string `json:"strategy,omitempty"`
	Trace    bool   `json:"trace,omitempty"`
}
