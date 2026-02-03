package yaml

// Proxy provides network proxy settings.
type Proxy struct {
	Url          string      `json:"url,omitempty"`
	Username     string      `json:"username,omitempty"`
	Password     string      `json:"password,omitempty"`
	ExcludeHosts interface{} `json:"excludeHosts,omitempty"`
}
