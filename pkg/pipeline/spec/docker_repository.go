package yaml

type DockerRepository struct {
	ConnectorRef string   `json:"connectorRef,omitempty"`
	Name         string   `json:"name,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}
