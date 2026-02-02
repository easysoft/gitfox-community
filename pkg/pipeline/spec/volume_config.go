package yaml

type VolumeConfigMap struct {
	Name     string `json:"name,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}
