package yaml

// Cache defines pipeline step caching behavior.
type Caching struct {
	Enabled *bool    `json:"enabled,omitempty"`
	Paths   []string `json:"paths,omitempty"`
	KeyFile string   `json:"keyFile,omitempty"`
}
