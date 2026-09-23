// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package ocischema

import (
	"encoding/json"

	"github.com/easysoft/gitfox-community/app/artifact/adapter/container/schema"

	"github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	unmarshalFn := func(b []byte) (schema.Manifest, error) {
		i := new(DeserializedIndex)
		if err := i.UnmarshalJSON(b); err != nil {
			return Manifest{}, err
		}
		return i, nil
	}
	err := schema.Register(v1.MediaTypeImageIndex, unmarshalFn)
	if err != nil {
		panic(err)
	}
}

var _ schema.Manifest = (*Index)(nil)

type Index struct {
	specs.Versioned
	MediaType string `json:"mediaType,omitempty"`
	Manifests []schema.Descriptor
}

func (i Index) GetMediaType() string {
	return i.MediaType
}

func (i Index) GetKind() schema.Kind {
	return schema.KindImageIndex
}

func (i Index) References() []schema.Descriptor { return i.Manifests }

type DeserializedIndex struct {
	Index
}

func (d *DeserializedIndex) UnmarshalJSON(b []byte) error {
	var obj Index
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	d.Index = obj
	return nil
}
