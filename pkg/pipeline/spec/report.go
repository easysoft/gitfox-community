package yaml

// Report defines a report artifact.
type Report struct {
	Path Stringorslice `json:"path,omitempty"`
	Type string        `json:"type,omitempty"`
}
