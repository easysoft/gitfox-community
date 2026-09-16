package gc

import (
	"context"
	"fmt"
	"slices"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	"github.com/easysoft/gitfox-community/app/artifact/adapter/container/schema"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	"github.com/easysoft/gitfox-community/pkg/storage"

	"github.com/opencontainers/go-digest"
)

// markManifestReferences recursively marks all references in a manifest tree.
// 递归地标记清单树中的所有引用。
func markManifestReferences(ctx context.Context, dgst digest.Digest, manifestService *manifestStore, ingester func(digest.Digest)) error {
	manifest, err := manifestService.Get(ctx, dgst)
	if err != nil {
		return fmt.Errorf("failed to retrieve manifest for digest %v: %w", dgst, err)
	}

	for _, descriptor := range manifest.References() {
		ingester(descriptor.Digest)
		if slices.Contains(schema.MediaList, descriptor.MediaType) {
			// is manifest, recurse mark
			err = markManifestReferences(ctx, descriptor.Digest, manifestService, ingester)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

type manifestStore struct {
	artStore  artifactStore.Interface
	fileStore storage.ContentStorage

	blobGetFn func(dgst digest.Digest) (*model.ArtifactAssetExtendBlob, bool)
}

// Get retrieves a manifest from the store.
// 从存储中检索清单。
func (m *manifestStore) Get(ctx context.Context, dgst digest.Digest) (schema.Manifest, error) {
	blob, ok := m.blobGetFn(dgst)
	if !ok {
		return nil, fmt.Errorf("digest not found")
	}
	content, err := m.fileStore.Get(ctx, adapter.BlobPath(blob.Ref))
	if err != nil {
		return nil, err
	}
	return schema.UnmarshalManifest(blob.ContentType, content)
}
