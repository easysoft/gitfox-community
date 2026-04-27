package adapter

import "encoding/json"

type MetadataInterface interface {
	ToJSON() (json.RawMessage, error)
}

type VersionMetadata struct {
	CreatorName string `json:"creator_name"`
}

func (m *VersionMetadata) ToJSON() (json.RawMessage, error) {
	return json.Marshal(m)
}
