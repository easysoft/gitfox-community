package yaml

// Mount defines a volume mount.
type Mount struct {
	Name     string `json:"name,omitempty"`
	Path     string `json:"path,omitempty"`
	SubPath  string `json:"subPath,omitempty"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}
