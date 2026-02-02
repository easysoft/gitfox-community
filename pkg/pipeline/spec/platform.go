package yaml

// Platform defines the target execution environment.
type Platform struct {
	Arch     string   `json:"arch,omitempty"`
	Features []string `json:"features,omitempty"`
	Os       string   `json:"os,omitempty"`
	Variant  string   `json:"variant,omitempty"`
	Version  string   `json:"version,omitempty"`
}
