package yaml

// Reference defines a git clone reference.
type Reference struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	Sha  string `json:"sha,omitempty"`
}
