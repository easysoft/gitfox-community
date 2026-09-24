// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package schema

import (
	"errors"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	DefaultOS   = "linux"
	DefaultArch = "amd64"
)

type Descriptor struct {
	// MediaType describe the type of the content. All text based formats are
	// encoded as utf-8.
	MediaType string `json:"mediaType,omitempty"`

	// Digest uniquely identifies the content. A byte stream can be verified
	// against this digest.
	Digest digest.Digest `json:"digest,omitempty"`

	// Size in bytes of content.
	Size int64 `json:"size,omitempty"`

	// URLs contains the source URLs of this content.
	URLs []string `json:"urls,omitempty"`

	// Annotations contains arbitrary metadata relating to the targeted content.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Platform describes the platform which the image in the manifest runs on.
	// This should only be used when referring to a manifest.
	Platform *v1.Platform `json:"platform,omitempty"`
}

type Kind string

const (
	KindImage      Kind = "image"
	KindImageIndex Kind = "index"
)

type Manifest interface {
	GetMediaType() string
	GetKind() Kind
	References() []Descriptor
}

type ImageManifest interface {
	Manifest
	GetLayers() []Descriptor
	// GetConfig 返回 image config blob 描述符（v2 schema 2 / OCI manifest v1 都有 config 字段）。
	// 用于读取 config blob 解析真实的 OS / Architecture，避免单平台 manifest 默认回落到 amd64。
	GetConfig() Descriptor
}

func ToImage(in Manifest) (ImageManifest, error) {
	if in.GetKind() != KindImage {
		return nil, errors.New("invalid kind")
	}
	im, ok := in.(ImageManifest)
	if !ok {
		return nil, errors.New("invalid image manifest")
	}
	return im, nil
}
