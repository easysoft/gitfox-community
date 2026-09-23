// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package schema

import (
	"fmt"
)

var schemaMap = make(map[string]UnmarshalFunc)

var MediaList = make([]string, 0)

type UnmarshalFunc func([]byte) (Manifest, error)

func Register(mediaType string, fn UnmarshalFunc) error {
	if _, ok := schemaMap[mediaType]; ok {
		return fmt.Errorf("manifest media type '%s' is already registered", mediaType)
	}
	schemaMap[mediaType] = fn
	MediaList = append(MediaList, mediaType)
	return nil
}

func UnmarshalManifest(mediaType string, b []byte) (Manifest, error) {
	fn, ok := schemaMap[mediaType]
	if !ok {
		return nil, fmt.Errorf("unsupported manifest media type '%s'", mediaType)
	}
	return fn(b)
}
