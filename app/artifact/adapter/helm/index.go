// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	"github.com/easysoft/gitfox-community/app/artifact/adapter/request"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	artifactStore "github.com/easysoft/gitfox-community/app/store/artifact"
	"github.com/easysoft/gitfox-community/pkg/storage"

	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"
)

const (
	_chartMetaType = "application/x-yaml"
)

type helmIndex struct {
	artStore  artifactStore.Interface
	fileStore storage.ContentStorage

	view     *adapter.ViewDescriptor
	indexReq *request.ArtifactIndexRequest

	descriptor *adapter.PackageMetaDescriptor
}

func EmptyIndex(contextPath string) []byte {
	index := NewIndexFile()
	index.ServerInfo = map[string]interface{}{
		"contextPath": contextPath,
	}
	content, _ := yaml.Marshal(index)
	return content
}

func NewHelmIndex(artStore artifactStore.Interface, view *adapter.ViewDescriptor) *helmIndex {
	return &helmIndex{artStore: artStore, fileStore: view.Store,
		view:       view,
		indexReq:   request.NewIndex(view, artStore),
		descriptor: adapter.NewEmptyPackageMetaDescriptor(),
	}
}

func (h *helmIndex) UpdatePackage(ctx context.Context, p *model.ArtifactPackage, v *model.ArtifactView) error {
	// Helm does not require implementation of package-level index
	return nil
}

func (h *helmIndex) UpdateRepo(ctx context.Context, contextPath string) error {
	index, err := h.buildIndex(ctx, contextPath)
	if err != nil {
		return errors.Wrap(err, "build helm index failed")
	}

	content, err := yaml.Marshal(index)
	if err != nil {
		return err
	}

	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.fileStore)
	if err != nil {
		return err
	}

	h.indexReq.RegisterWriter(fw)
	size, hash, err := request.Write(bytes.NewReader(content), fw)
	if err != nil {
		return errors.Wrap(err, "write helm index failed")
	}

	logger.GetLogger(ctx).Infof("generate helm index: %s", string(content))

	h.descriptor.Format = model.ArtifactHelmFormat
	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Path = "index.yaml"
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = _chartMetaType

	h.indexReq.Descriptor = h.descriptor

	if err = h.indexReq.Commit(ctx); err != nil {
		logger.GetLogger(ctx).Errorf("commit index failed: %v", err)
		return h.indexReq.Cancel(ctx)
	}
	return nil
}

func (h *helmIndex) buildIndex(ctx context.Context, contextPath string) (*IndexFile, error) {
	log := logger.GetLogger(ctx)

	index := NewIndexFile()

	// setup dynamic contextPath for cm-push command tool
	// https://github.com/chartmuseum/helm-push/blob/main/pkg/chartmuseum/upload.go#L22
	index.ServerInfo = map[string]interface{}{
		"contextPath": contextPath,
	}
	packages, err := h.artStore.Packages().List(ctx, h.view.ArtifactID, false)
	if err != nil {
		return nil, err
	}

	var items []*model.ArtifactVersion
	for _, pkg := range packages {
		if pkg.Format != model.ArtifactHelmFormat {
			continue
		}
		vers, e := h.artStore.Versions().Find(ctx, model.SearchVersionOption{PackageID: pkg.ID})
		if e != nil {
			log.Errorf("find versions for package %d failed: %v", pkg.ID, e)
			return nil, e
		}
		items = append(items, vers...)
	}

	log.Infof("find items: %+v", items)

	for _, item := range items {
		assets, e := h.artStore.Assets().Search(ctx,
			model.SearchAssetOption{Rank: model.AssetRankMain, VersionID: item.ID},
			artifactStore.AssetExcludeDeletedOption{})
		if e != nil {
			log.Errorf("search assets failed: %v", e)
			return nil, e
		}

		if len(assets) < 1 {
			log.Errorf("ignore version with null asset of %s", item.Version)
			continue
		}

		asset := assets[0]

		var hash adapter.Hash
		if e = json.Unmarshal([]byte(asset.CheckSum), &hash); e != nil {
			log.Errorf("ignore invalid checksum asset of %s", item.Version)
			continue
		}

		if asset.Metadata == "" {
			continue
		}
		var meta chart.Metadata
		if e = json.Unmarshal([]byte(asset.Metadata), &meta); e != nil {
			log.Errorf("unmarshal helm metadata failed: %v", e)
			log.Warnf("ignore invalid helm metadata for %s", item.Version)
			continue
		}

		if e = index.MustAdd(&meta, asset.Path, "", hash.Sha256); e != nil {
			log.Errorf("add chart entry failed: %v", e)
			continue
		}
	}
	index.SortEntries()
	return index, nil
}

// IndexFile represents the index file in a chart repository
type IndexFile struct {
	// This is used ONLY for validation against chartmuseum's index files and is discarded after validation.
	ServerInfo map[string]interface{}   `json:"serverInfo,omitempty"`
	APIVersion string                   `json:"apiVersion"`
	Generated  time.Time                `json:"generated"`
	Entries    map[string]ChartVersions `json:"entries"`
	PublicKeys []string                 `json:"publicKeys,omitempty"`

	// Annotations are additional mappings uninterpreted by Helm. They are made available for
	// other applications to add information to the index file.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// APIVersionV1 is the v1 API version for index and repository files.
const APIVersionV1 = "v1"

// NewIndexFile initializes an index.
func NewIndexFile() *IndexFile {
	return &IndexFile{
		APIVersion: APIVersionV1,
		Generated:  time.Now(),
		Entries:    map[string]ChartVersions{},
		PublicKeys: []string{},
	}
}

// MustAdd adds a file to the index
// This can leave the index in an unsorted state
func (i IndexFile) MustAdd(md *chart.Metadata, filename, baseURL, digest string) error {
	if i.Entries == nil {
		return errors.New("entries not initialized")
	}

	if md.APIVersion == "" {
		md.APIVersion = chart.APIVersionV1
	}
	if err := md.Validate(); err != nil {
		return errors.Wrapf(err, "validate failed for %s", filename)
	}

	u := filename
	if baseURL != "" {
		_, file := filepath.Split(filename)
		var err error
		u, err = URLJoin(baseURL, file)
		if err != nil {
			u = path.Join(baseURL, file)
		}
	}
	cr := &ChartVersion{
		URLs:     []string{u},
		Metadata: md,
		Digest:   digest,
		Created:  time.Now(),
	}
	ee := i.Entries[md.Name]
	i.Entries[md.Name] = append(ee, cr)
	return nil
}

// SortEntries sorts the entries by version in descending order.
//
// In canonical form, the individual version records should be sorted so that
// the most recent release for every version is in the 0th slot in the
// Entries.ChartVersions array. That way, tooling can predict the newest
// version without needing to parse SemVers.
func (i IndexFile) SortEntries() {
	for _, versions := range i.Entries {
		sort.Sort(sort.Reverse(versions))
	}
}

// URLJoin joins a base URL to one or more path components.
//
// It's like filepath.Join for URLs. If the baseURL is pathish, this will still
// perform a join.
//
// If the URL is unparsable, this returns an error.
func URLJoin(baseURL string, paths ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	// We want path instead of filepath because path always uses /.
	all := []string{u.Path}
	all = append(all, paths...)
	u.Path = path.Join(all...)
	return u.String(), nil
}
