// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package gc

import (
	"context"
	"fmt"
	"time"

	// Register Docker schema2 and OCI manifest decoders for standalone GC use
	// and tests that do not import the container controller package.
	_ "github.com/easysoft/gitfox-community/app/artifact/adapter/container/schema/ocischema"
	_ "github.com/easysoft/gitfox-community/app/artifact/adapter/container/schema/schema2"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	"github.com/easysoft/gitfox-community/pkg/storage"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"

	"github.com/opencontainers/go-digest"
)

type digestRef struct {
	count int
	obj   *model.ArtifactAssetExtendBlob
}

type manifestRefs struct {
	digests []digest.Digest
	obj     *model.ArtifactAssetExtendBlob
}

type TagDescriptor struct {
	AssetId       int64
	VersionId     int64
	ExclusiveSize int64
	TotalSize     int64
	ExcludeRefs   int
	TotalRefs     int
}

type StatisticReport struct {
	DeleteList []*model.ArtifactRecycleBlobDesc
	TagList    []*TagDescriptor
}

// Statistic builds container capacity and GC candidates from the OCI/Docker
// manifest graph. Container blob liveness is defined by digests reachable from
// active tag manifests, not by simple asset.blobID reference counts.
func Statistic(ctx context.Context, artStore artifactStore.Interface, fileStore storage.ContentStorage, retentionTime time.Duration) (*model.ArtifactStatisticReport, error) {
	// 1. 获取所有容器格式的资产 (含 manifest, config, layer)
	allAssets, err := artStore.Assets().SearchExtendBlob(ctx, ContainerFormatOption{})
	if err != nil {
		return nil, err
	}

	// 2. 构建全局 digest 映射表，初始引用计数 count 置为 0
	digestRefsMap := make(map[digest.Digest]*digestRef, len(allAssets))
	for _, asset := range allAssets {
		if asset == nil {
			continue
		}
		dgst, err := digest.Parse(asset.Path)
		if err != nil {
			return nil, err
		}
		digestRefsMap[dgst] = &digestRef{count: 0, obj: asset}
	}

	// 3. 构建 manifest 读取器，用于后续从存储中解析 manifest 的子引用记录
	manifestGetter := &manifestStore{
		artStore:  artStore,
		fileStore: fileStore,
		blobGetFn: func(dgst digest.Digest) (*model.ArtifactAssetExtendBlob, bool) {
			ref, ok := digestRefsMap[dgst]
			if !ok {
				return nil, false
			}
			return ref.obj, true
		},
	}

	// 4. 获取作为 GC 根节点的 tag (包括活跃的，以及在回收保留期内的软删 tag)
	tagSearchOpts := []artifactStore.SearchOption{TagAssetOption{}}
	if retentionTime > 0 {
		tagSearchOpts = append(tagSearchOpts, TagUndeletedAndRecentDeletedOption{After: datetime.NowLocalTime().Add(-retentionTime).Time()})
	} else {
		tagSearchOpts = append(tagSearchOpts, TagActiveOption{})
	}
	taggedAssets, err := artStore.Assets().SearchExtendBlob(ctx, tagSearchOpts...)
	if err != nil {
		return nil, err
	}

	// 5. 遍历 GC 根节点 (tag)，递归遍历其 manifest 引用树，标记可达的 digest (count++)
	assetMarkMap := make(map[int64]*manifestRefs, len(taggedAssets))
	for _, tagAsset := range taggedAssets {
		if tagAsset == nil {
			continue
		}
		dgst, err := digest.Parse(tagAsset.Path)
		if err != nil {
			return nil, err
		}
		if _, ok := digestRefsMap[dgst]; !ok {
			return nil, fmt.Errorf("unknown tag asset: %s", dgst.String())
		}
		digestRefsMap[dgst].count++

		manifestRefObj := &manifestRefs{digests: make([]digest.Digest, 0), obj: tagAsset}
		assetMarkMap[tagAsset.ID] = manifestRefObj
		manifestRefObj.digests = append(manifestRefObj.digests, dgst)

		err = markManifestReferences(ctx, dgst, manifestGetter, func(d digest.Digest) {
			_, exist := digestRefsMap[d]
			if !exist {
				return
			}
			// 累加子节点的引用计数
			digestRefsMap[d].count += 1
			manifestRefObj.digests = append(manifestRefObj.digests, d)
		})
		if err != nil {
			return nil, err
		}

	}

	// 6. 统计每个 tag 的容量，区分独占容量 (Exclusive) 和总容量 (Total)
	tagList := make([]*model.ArtifactVersionCapacityDesc, 0, len(assetMarkMap))
	for _, assetMark := range assetMarkMap {
		item := &model.ArtifactVersionCapacityDesc{
			AssetID:   assetMark.obj.ID,
			VersionID: assetMark.obj.VersionID.Int64,
		}

		for _, ref := range assetMark.digests {
			d, ok := digestRefsMap[ref]
			if !ok {
				continue
			}
			if d.count == 1 {
				item.ExclusiveRefs++
				item.ExclusiveSize += d.obj.Size
			} else if d.count > 1 {
				d.count--
			}
			item.TotalRefs++
			item.TotalSize += d.obj.Size
		}
		tagList = append(tagList, item)
	}

	// 7. 找出待清理的资产。有两种情况需要清理：
	//    a. 整个 digest 不可达 (count == 0)，则该 digest 下的所有资产都要硬删，且附带物理删除 (DeleteBlob = true)
	//    b. 整个 digest 依然可达 (count > 0)，但某些特定的 tag 资产已过保留期，
	//       这些过期的资产行应当从数据库中清理，但底层 blob 不能物理删除 (DeleteBlob = false)。
	deleteList := make([]*model.ArtifactRecycleBlobDesc, 0)
	retentionCutoff := datetime.NowLocalTime().Add(-retentionTime)

	for _, asset := range allAssets {
		if asset == nil {
			continue
		}
		dgst, err := digest.Parse(asset.Path)
		if err != nil {
			continue
		}
		d, ok := digestRefsMap[dgst]
		if !ok {
			continue
		}

		if d.count == 0 {
			blobRef := asset.Ref
			deleteList = append(deleteList, &model.ArtifactRecycleBlobDesc{
				AssetID:    asset.ID,
				VersionID:  asset.VersionID.Int64,
				GroupID:    asset.GroupID.Int64,
				ArtifactID: asset.ArtifactID.Int64,
				Path:       dgst.String(),
				BlobID:     asset.BlobID,
				BlobRef:    blobRef,
				DeleteBlob: true,
				Size:       asset.Size,
			})
		} else {
			isTag := asset.VersionID.Int64 > 0
			if isTag && asset.Deleted > 0 && asset.EditedDate != nil && asset.EditedDate.Before(*retentionCutoff) {
				deleteList = append(deleteList, &model.ArtifactRecycleBlobDesc{
					AssetID:    asset.ID,
					VersionID:  asset.VersionID.Int64,
					GroupID:    asset.GroupID.Int64,
					ArtifactID: asset.ArtifactID.Int64,
					Path:       dgst.String(),
					BlobID:     asset.BlobID,
					BlobRef:    asset.Ref,
					DeleteBlob: true,
					Size:       asset.Size,
				})
			}
		}
	}

	return &model.ArtifactStatisticReport{DeleteList: deleteList, TagList: tagList}, nil
}
