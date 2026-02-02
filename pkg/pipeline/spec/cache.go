package yaml

// Cache defines pipeline caching behavior.
type Cache struct {
	Enabled bool     `json:"enabled,omitempty"`
	Paths   []string `json:"paths,omitempty"`
	Key     string   `json:"key,omitempty"`
	Policy  string   `json:"policy,omitempty"`
}
