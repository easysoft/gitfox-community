// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package schema2

import (
	"encoding/json"

	"github.com/easysoft/gitfox-community/app/artifact/adapter/container/schema"

	"github.com/opencontainers/image-spec/specs-go"
)

const (
	MediaTypeManifest = "application/vnd.docker.distribution.manifest.v2+json"
)

func init() {
	unmarshalFn := func(b []byte) (schema.Manifest, error) {
		i := new(DeserializedManifest)
		if err := i.UnmarshalJSON(b); err != nil {
			return Manifest{}, err
		}
		return i, nil
	}
	err := schema.Register(MediaTypeManifest, unmarshalFn)
	if err != nil {
		panic(err)
	}
}

type Manifest struct {
	specs.Versioned
	MediaType string `json:"mediaType,omitempty"`

	// Config references the image configuration as a blob.
	Config schema.Descriptor `json:"config"`

	// Layers lists descriptors for the layers referenced by the
	// configuration.
	Layers []schema.Descriptor `json:"layers"`
}

// References returns the descriptors of this manifests references.
func (m Manifest) References() []schema.Descriptor {
	references := make([]schema.Descriptor, 0, 1+len(m.Layers))
	references = append(references, m.Config)
	references = append(references, m.Layers...)
	return references
}

func (m Manifest) GetMediaType() string {
	return m.MediaType
}

func (m Manifest) GetKind() schema.Kind {
	return schema.KindImage
}

func (m Manifest) GetLayers() []schema.Descriptor {
	return m.Layers
}

func (m Manifest) GetConfig() schema.Descriptor {
	return m.Config
}

type DeserializedManifest struct {
	Manifest
}

func (d *DeserializedManifest) UnmarshalJSON(b []byte) error {
	var obj Manifest
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	d.Manifest = obj
	return nil
}
