package yaml

// Defines the container resource request and limits.
type Resources struct {
	Limits   *Resource `json:"limits,omitempty"`
	Requests *Resource `json:"requests,omitempty"`
}
