package yaml

type VolumeTemp struct {
	Medium string      `json:"medium,omitempty"`
	Limit  StringorInt `json:"limit,omitempty"`
}
