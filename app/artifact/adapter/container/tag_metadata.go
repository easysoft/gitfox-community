// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package container

import (
	"context"
	"encoding/json"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	"github.com/easysoft/gitfox-community/app/artifact/adapter/container/schema"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	"github.com/easysoft/gitfox-community/pkg/logger"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type TagMetadataReader struct {
	artStore artifactStore.Interface
	view     *adapter.ViewDescriptor
}

type TagMetadata struct {
	Images []*ImageMeta `json:"images"`
}

type ImageMeta struct {
	Digest digest.Digest `json:"digest"`
	OS     string        `json:"os"`
	Arch   string        `json:"arch"`
	Size   int64         `json:"size"`
}

func NewTagMetadataReader(artStore artifactStore.Interface, view *adapter.ViewDescriptor) *TagMetadataReader {
	return &TagMetadataReader{artStore: artStore, view: view}
}

func (t *TagMetadataReader) Parse(ctx context.Context, mediaType string, digest digest.Digest) (*TagMetadata, error) {
	manifest, err := t.parseDigest(ctx, mediaType, digest)
	if err != nil {
		return nil, err
	}

	platforms := make([]*ImageMeta, 0)

	if manifest.GetKind() == schema.KindImageIndex {
		for _, reference := range manifest.References() {
			p := reference.Platform
			if p.OS == "unknown" {
				continue
			}

			nextManifest, e := t.parseDigest(ctx, reference.MediaType, reference.Digest)
			if e != nil {
				return nil, e
			}
			imageMani, e := t.parseImage(ctx, p, nextManifest)
			if e != nil {
				return nil, e
			}
			imageMani.Digest = reference.Digest

			platforms = append(platforms, imageMani)
		}
	} else {
		imgManifest, e := t.parseImage(ctx, nil, manifest)
		if e != nil {
			return nil, e
		}
		imgManifest.Digest = digest
		platforms = append(platforms, imgManifest)
	}

	return &TagMetadata{
		Images: platforms,
	}, nil
}

func (t *TagMetadataReader) parseImage(ctx context.Context, platform *v1.Platform, m schema.Manifest) (*ImageMeta, error) {
	im, err := schema.ToImage(m)
	if err != nil {
		return nil, err
	}

	var data = ImageMeta{}
	if platform != nil {
		// multi-arch index 路径：直接用 index 里的 platform 字段。
		data.OS = platform.OS
		data.Arch = platform.Architecture
	} else {
		// 单平台 manifest 路径：manifest 顶层不带 platform，需要从 config blob 解析真实 OS/Arch。
		// 解析失败时回落到 schema.DefaultOS / DefaultArch，仅作兜底，避免破坏现有列表。
		data.OS = schema.DefaultOS
		data.Arch = schema.DefaultArch
		if cfgOS, cfgArch, cfgErr := t.readImageConfigPlatform(ctx, im.GetConfig()); cfgErr == nil {
			if cfgOS != "" {
				data.OS = cfgOS
			}
			if cfgArch != "" {
				data.Arch = cfgArch
			}
		} else {
			logger.GetLogger(ctx).Errorf("【container metadata】read image config blob failed, fallback to default os/arch: %v", cfgErr)
		}
	}

	for _, layer := range im.GetLayers() {
		data.Size = data.Size + layer.Size
	}
	return &data, nil
}

// readImageConfigPlatform 从 image config blob 中解析 architecture / os 字段。
// config blob 的 JSON 结构遵循 OCI image-spec：{"architecture": "...", "os": "...", ...}。
func (t *TagMetadataReader) readImageConfigPlatform(ctx context.Context, cfg schema.Descriptor) (string, string, error) {
	if cfg.Digest == "" {
		return "", "", nil
	}
	dbAsset, err := t.artStore.Assets().GetPath(ctx, cfg.Digest.String(), model.ArtifactContainerFormat)
	if err != nil {
		return "", "", err
	}
	dbBlob, err := t.artStore.Blobs().GetById(ctx, dbAsset.BlobID)
	if err != nil {
		return "", "", err
	}
	b, err := t.view.Store.Get(ctx, adapter.BlobPath(dbBlob.Ref))
	if err != nil {
		return "", "", err
	}
	var img v1.Image
	if err = json.Unmarshal(b, &img); err != nil {
		return "", "", err
	}
	return img.OS, img.Architecture, nil
}

func (t *TagMetadataReader) parseDigest(ctx context.Context, mediaType string, digest digest.Digest) (schema.Manifest, error) {
	dbAsset, err := t.artStore.Assets().GetPath(ctx, digest.String(), model.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}

	dbBlob, err := t.artStore.Blobs().GetById(ctx, dbAsset.BlobID)
	if err != nil {
		return nil, err
	}

	b, err := t.view.Store.Get(ctx, adapter.BlobPath(dbBlob.Ref))
	if err != nil {
		return nil, err
	}

	return schema.UnmarshalManifest(mediaType, b)
}
