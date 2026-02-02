package yaml

// Clone defines the clone behavior.
type Clone struct {
	Depth    int64      `json:"depth,omitempty"`
	Disabled bool       `json:"disabled,omitempty"`
	Insecure bool       `json:"insecure,omitempty"`
	Strategy string     `json:"strategy,omitempty"`
	Trace    bool       `json:"trace,omitempty"`
	Ref      *Reference `json:"ref,omitempty"`
}
