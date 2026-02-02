package yaml

// Resource defines the system resources.
type Resource struct {
	Cpu    StringorInt    `json:"cpu,omitempty"`
	Memory MemStringorInt `json:"memory,omitempty"`
}
