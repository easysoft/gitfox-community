// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"archive/zip"
	"compress/flate"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/easysoft/gitfox-community/app/artifact/adapter"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/pkg/logger"
	storagedriver "github.com/easysoft/gitfox-community/pkg/storage/driver"
)

type ArchiveGroup struct {
	base      *archiveGroupItem
	subGroups []*archiveGroupItem
}

func (a *ArchiveGroup) Name() string {
	return fmt.Sprintf("%s.zip", a.base.Name)
}

type archiveGroupItem struct {
	*model.ArtifactGroup
	archivePath string
}

// PrepareArchiveGroup 准备归档分组
func (s *Service) PrepareArchiveGroup(ctx context.Context, session *authModel.Session, groupID int64) (*ArchiveGroup, error) {
	root, err := s.artStore.Groups().GetById(ctx, groupID)
	if err != nil {
		return nil, err
	}

	descendants, err := s.artStore.Groups().ListDescendants(ctx, groupID)
	if err != nil {
		return nil, err
	}
	rootGroup := &archiveGroupItem{ArtifactGroup: root, archivePath: root.Name}
	parentMap := map[int64]*archiveGroupItem{groupID: rootGroup}
	archive := &ArchiveGroup{base: rootGroup}
	for _, d := range descendants {
		parent := parentMap[d.ParentID.Int64]
		if parent == nil {
			logger.GetLogger(ctx).Errorf("parent group %d does not exist in archive parent map", d.ParentID.Int64)
			continue
		}
		item := &archiveGroupItem{ArtifactGroup: d, archivePath: filepath.Join(parent.archivePath, d.Name)}
		parentMap[d.ID] = item
		archive.subGroups = append(archive.subGroups, item)
	}
	return archive, nil
}

// ArchiveGroup 归档文件夹
func (s *Service) ArchiveGroup(ctx context.Context, archive *ArchiveGroup, output io.Writer) error {
	zipWriter := zip.NewWriter(output)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	defer zipWriter.Close()

	for _, g := range append(archive.subGroups, archive.base) {
		assets, err := s.artStore.ListGroupAssets(ctx, g.ID)
		if err != nil {
			return err
		}
		for _, asset := range assets {
			info, err := s.fileStore.Stat(ctx, adapter.BlobPath(asset.BlobRef))
			if err != nil {
				return err
			}
			zf := zipAssetFile{asset: asset, info: info}
			header, err := zip.FileInfoHeader(zf)
			if err != nil {
				return err
			}
			header.Name = filepath.Join(g.archivePath, asset.Path)
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			reader, err := s.fileStore.Open(ctx, adapter.BlobPath(asset.BlobRef))
			if err != nil {
				return err
			}
			if _, err = io.Copy(writer, reader); err != nil {
				_ = reader.Close()
				return err
			}
			_ = reader.Close()
		}
	}
	return zipWriter.Flush()
}

type zipAssetFile struct {
	asset *model.ArtifactAssetsRes
	info  storagedriver.FileInfo
}

func (z zipAssetFile) Name() string       { return z.asset.Path }
func (z zipAssetFile) Size() int64        { return z.info.Size() }
func (z zipAssetFile) Mode() fs.FileMode  { return 0644 }
func (z zipAssetFile) ModTime() time.Time { return z.info.ModTime() }
func (z zipAssetFile) IsDir() bool        { return z.info.IsDir() }
func (z zipAssetFile) Sys() any           { return nil }
