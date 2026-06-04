package gc

import (
	"context"
	"fmt"
	"time"

	// Register Docker schema2 and OCI manifest decoders for standalone GC use
	// and tests that do not import the container controller package.
	_ "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container/schema/ocischema"
	_ "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container/schema/schema2"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"

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
	allAssets, err := artStore.Assets().SearchExtendBlob(ctx, ContainerFormatOption{})
	if err != nil {
		return nil, err
	}

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

	tagSearchOpts := []artifactStore.SearchOption{TagAssetOption{}}
	if retentionTime > 0 {
		tagSearchOpts = append(tagSearchOpts, TagUndeletedAndRecentDeletedOption{After: time.Now().Add(-retentionTime)})
	} else {
		tagSearchOpts = append(tagSearchOpts, TagActiveOption{})
	}
	taggedAssets, err := artStore.Assets().SearchExtendBlob(ctx, tagSearchOpts...)
	if err != nil {
		return nil, err
	}

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
			digestRefsMap[d].count += 1
			manifestRefObj.digests = append(manifestRefObj.digests, d)
		})
		if err != nil {
			return nil, err
		}

	}

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

	deleteList := make([]*model.ArtifactRecycleBlobDesc, 0)
	for dgst, d := range digestRefsMap {
		if d.count != 0 {
			continue
		}
		asset := d.obj
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

	return &model.ArtifactStatisticReport{DeleteList: deleteList, TagList: tagList}, nil
}
