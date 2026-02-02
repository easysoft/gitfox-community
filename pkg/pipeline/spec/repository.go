package yaml

// Repository defines a remote git repository.
type Repository struct {
	Connector string `json:"connector,omitempty"`
	Name      string `json:"name,omitempty"`
}
