package yaml

type VolumeClaim struct {
	ClaimName string `json:"claimName,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}
