package container

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	ctnAdapter "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"

	"github.com/opencontainers/go-digest"
)

const (
	headerContentType   = "Content-Type"
	headerContentLength = "Content-Length"
	headerDisposition   = "Content-Disposition"
	headerCacheControl  = "Cache-Control"
	headerEtag          = "ETag"
	headerLastModify    = "Last-Modified"
	headerLocation      = "Location"
	headerRange         = "Range"
	headerDigest        = "Docker-Content-Digest"
	HeadDockerUUID      = "Docker-Upload-UUID"
)

const (
	blobCache = time.Hour * 24 * 365
)

var regexTag = regexp.MustCompile("^[a-zA-Z0-9_][a-zA-Z0-9._-]{0,127}$")

func validateTag(s string) bool {
	return regexTag.MatchString(s)
}

// Controller 容器镜像（OCI v2 协议）控制器。
// 负责处理 docker login / pull / push 的核心业务逻辑。
type Controller struct {
	artStore  artifactStore.Interface
	fileStore storage.ContentStorage
}

// NewController 创建容器镜像控制器实例。
func NewController(
	artStore artifactStore.Interface,
	fileStore storage.ContentStorage,
) *Controller {
	return &Controller{
		artStore:  artStore,
		fileStore: fileStore,
	}
}

// ---- 请求解析 ----

// ContainerReq 容器镜像请求的基础参数。
type ContainerReq struct {
	// Scope 作用域：global | space | repo，由路由前缀决定。
	Scope string
	// SpaceName 空间标识，scope=space/repo 时有值。
	SpaceName string
	// ScopeRepo 仓库标识（git repo identifier），scope=repo 时有值。
	ScopeRepo string
	// ViewName 制品视图名称，在 scope 作用域内唯一。
	ViewName string
	// GroupName container namespace（必填），同 view 内唯一。
	GroupName string
	// RepoName 镜像名（container image name）。
	RepoName string
	View     *adapter.ViewDescriptor
}

// FullName 返回 OCI repository name，即 URL 中 /v2/ 之后到操作路径之前的部分。
// 格式与 docker client 使用的路径完全一致，用作 blob upload session 的唯一 key。
//
//	global: "global/{viewname}/{groupname}/{imagename}"
//	space:  "space/{spaceName}/{viewname}/{groupname}/{imagename}"
//	repo:   "repo/{spaceName}/{scopeRepo}/{viewname}/{groupname}/{imagename}"
func (cr *ContainerReq) FullName() string {
	switch cr.Scope {
	case "global":
		return "global/" + cr.ViewName + "/" + cr.GroupName + "/" + cr.RepoName
	case "space":
		return "space/" + cr.SpaceName + "/" + cr.ViewName + "/" + cr.GroupName + "/" + cr.RepoName
	default: // repo
		return "repo/" + cr.SpaceName + "/" + cr.ScopeRepo + "/" + cr.ViewName + "/" + cr.GroupName + "/" + cr.RepoName
	}
}

// ContainerManifestRequest 容器 manifest 请求。
type ContainerManifestRequest struct {
	*ContainerReq
	IsTag  bool
	Tag    string
	Digest digest.Digest
}

// ContainerUploadRequest 容器 blob 上传续传请求。
type ContainerUploadRequest struct {
	*ContainerReq
	Resume *ctnAdapter.ResumeRequest
}

// ParseManifestReference 将 reference 解析为 tag 或 digest。
func ParseManifestReference(base *ContainerReq, reference string) (*ContainerManifestRequest, error) {
	req := &ContainerManifestRequest{ContainerReq: base}
	if validateTag(reference) {
		req.Tag = reference
		req.IsTag = true
		return req, nil
	}
	dg, err := digest.Parse(reference)
	if err != nil {
		return nil, ctnAdapter.ErrManifestUnknown.WithDetail(err.Error())
	}
	req.Digest = dg
	return req, nil
}

// ---- AssetMeta ----

// AssetMeta 制品资产的元数据（用于 blob/manifest 读取）。
type AssetMeta struct {
	Id          int64
	BlobId      int64
	Size        int64
	Path        string
	ContentType string
	ETag        string
	LastModify  *time.Time
	Ref         string
}

func newAssetMeta(asset *model.ArtifactAsset, blob *model.ArtifactBlob) *AssetMeta {
	return &AssetMeta{
		Id:          asset.ID,
		BlobId:      blob.ID,
		Size:        blob.Size,
		Path:        asset.Path,
		ContentType: asset.ContentType,
		ETag:        "",
		LastModify:  blob.CreatedDate,
		Ref:         blob.Ref,
	}
}

func (r *AssetMeta) Write(w http.ResponseWriter) {
	w.Header().Set(headerContentLength, fmt.Sprintf("%d", r.Size))
	w.Header().Set(headerContentType, r.ContentType)
	if r.ETag != "" {
		w.Header().Set(headerEtag, r.ETag)
	}
	if r.LastModify != nil {
		w.Header().Set(headerLastModify, r.LastModify.Format(time.RFC1123))
	}
	w.Header().Set(headerDisposition, fmt.Sprintf("filename=%s", filepath.Base(r.Path)))
}

// ---- Asset 读取 ----

// GetAssetInfo 根据 path 和 format 获取制品资产元信息。
func (c *Controller) GetAssetInfo(ctx context.Context, path string, format model.ArtifactFormat) (*AssetMeta, error) {
	asset, err := c.artStore.Assets().GetPath(ctx, path, format)
	if err != nil {
		return nil, err
	}
	blob, err := c.artStore.Blobs().GetById(ctx, asset.BlobID)
	if err != nil {
		return nil, err
	}
	return newAssetMeta(asset, blob), nil
}

// GetAssetReader 获取制品资产的文件流。
func (c *Controller) GetAssetReader(ctx context.Context, path string, format model.ArtifactFormat) (io.ReadCloser, *AssetMeta, error) {
	meta, err := c.GetAssetInfo(ctx, path, format)
	if err != nil {
		return nil, nil, err
	}
	r, err := c.fileStore.Open(ctx, adapter.BlobPath(meta.Ref))
	return r, meta, err
}

// ---- 工具方法 ----

// handleUpload 通用上传处理：Serve → Validate → Save。
func handleUpload(ctx context.Context, req *http.Request, upload adapter.ArtifactPackageUploader) error {
	if _, err := upload.Serve(ctx, req); err != nil {
		return err
	}
	if err := upload.IsValid(ctx); err != nil {
		_ = upload.Cancel(ctx)
		return err
	}
	return upload.Save(ctx)
}

// ---- tagInfo ----

type tagInfo struct {
	name    string
	version *model.ArtifactVersion
	asset   *model.ArtifactAsset
	blob    *model.ArtifactBlob
}

// findContainerTag 按镜像名、命名空间（groupName）和 tag 查找镜像版本及关联资源。
func findContainerTag(ctx context.Context, artStore artifactStore.Interface, viewId int64, pkgName, groupName, tag string) (*tagInfo, error) {
	log := logger.GetLogger(ctx)
	dbPkg, err := artStore.Packages().GetByName(ctx, pkgName, groupName, viewId, model.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}
	dbVer, err := artStore.Versions().GetByVersion(ctx, dbPkg.ID, viewId, tag)
	if err != nil {
		return nil, err
	}
	dbAssets, err := artStore.Assets().FindMain(ctx, dbVer.ID)
	if err != nil {
		return nil, err
	}
	if len(dbAssets) != 1 {
		log.Errorf("expected 1 asset for tag %s, found %d", tag, len(dbAssets))
		return nil, fmt.Errorf("expected 1 asset, found %d", len(dbAssets))
	}
	dbAsset := dbAssets[0]
	dbBlob, err := artStore.Blobs().GetById(ctx, dbAsset.BlobID)
	if err != nil {
		return nil, err
	}
	return &tagInfo{
		name:    tag,
		version: dbVer,
		asset:   dbAsset,
		blob:    dbBlob,
	}, nil
}

// ---- Blob 操作 ----

// HeadBlob HEAD /v2/{space}/{name}/blobs/{digest}
func (c *Controller) HeadBlob(ctx context.Context, ctnReq *ContainerReq, dgst string) (HttpResponseWriter, error) {
	return c.getContainerBlob(ctx, ctnReq, dgst, true)
}

// GetContainerBlob GET /v2/{space}/{name}/blobs/{digest}
func (c *Controller) GetContainerBlob(ctx context.Context, ctnReq *ContainerReq, dgst string) (HttpResponseWriter, error) {
	return c.getContainerBlob(ctx, ctnReq, dgst, false)
}

func (c *Controller) getContainerBlob(ctx context.Context, _ *ContainerReq, dgst string, head bool) (HttpResponseWriter, error) {
	// TODO: 权限检查（pull 需要至少读权限，先空实现）
	fr, meta, err := c.GetAssetReader(ctx, dgst, model.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}
	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Set(headerEtag, fmt.Sprintf(`"%s"`, meta.Path))
		w.Header().Set(headerCacheControl, fmt.Sprintf("max-age=%.f", blobCache.Seconds()))
		w.Header().Add(headerContentType, meta.ContentType)
		w.Header().Add(headerContentLength, fmt.Sprintf("%d", meta.Size))
		w.Header().Add(headerDigest, meta.Path)
		defer fr.Close()
		if !head {
			renderReader(ctx, w, http.StatusOK, fr)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}), nil
}

// DeleteContainerBlob DELETE /v2/{space}/{name}/blobs/{digest}
func (c *Controller) DeleteContainerBlob(ctx context.Context, _ *ContainerReq, dgst string) (HttpResponseWriter, error) {
	// TODO: 权限检查（delete 需要读写权限，先空实现）
	meta, err := c.GetAssetInfo(ctx, dgst, model.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}
	if err = c.artStore.Blobs().DeleteById(ctx, meta.BlobId); err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			return nil, ctnAdapter.ErrManifestBlobUnknown.WithDetail(err.Error())
		}
		return nil, ctnAdapter.ErrUnknown.WithDetail(err.Error())
	}
	if err = c.artStore.Assets().DeleteById(ctx, meta.Id); err != nil {
		return nil, ctnAdapter.ErrUnknown.WithDetail(err.Error())
	}
	_ = c.fileStore.Delete(ctx, adapter.BlobPath(meta.Ref))

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusAccepted)
	}), nil
}

// ---- Blob Upload 操作 ----

// SaveContainerBlob 单次上传 blob（带 digest 参数时使用）。
func (c *Controller) SaveContainerBlob(ctx context.Context, ctnReq *ContainerReq, req *http.Request, reference digest.Digest) (HttpResponseWriter, error) {
	// TODO: 权限检查（push 需要读写权限，先空实现）
	refDigest := reference.String()
	uploader := ctnAdapter.NewManifestUploader(c.artStore, ctnReq.View, refDigest)

	if err := handleUpload(ctx, req, uploader); err != nil {
		return nil, err
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerContentLength, "0")
		w.WriteHeader(http.StatusCreated)
	}), nil
}

// StartBlobUpload 启动分块 blob 上传会话。
func (c *Controller) StartBlobUpload(ctx context.Context, ctnReq *ContainerReq) (HttpResponseWriter, error) {
	// TODO: 权限检查（push 需要读写权限，先空实现）
	req := ctnAdapter.InitResumeRequest(ctnReq.FullName(), c.artStore, ctnReq.View)
	return c.getUploadState(ctx, ctnReq, req)
}

func (c *Controller) getUploadState(ctx context.Context, ctnReq *ContainerReq, req *ctnAdapter.ResumeRequest) (HttpResponseWriter, error) {
	token, err := req.State.Pack("")
	if err != nil {
		return nil, err
	}

	endSize, err := req.Size(ctx)
	if err != nil {
		endSize = 0
	}
	if endSize > 0 {
		endSize = endSize - 1
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(HeadDockerUUID, req.State.SessionID)
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerRange, fmt.Sprintf("0-%d", endSize))
		// 将 _state token 作为 query 参数传回
		uploadURL := fmt.Sprintf("/v2/%s/blobs/uploads/%s?_state=%s", ctnReq.FullName(), req.SessionID, token)
		w.Header().Add(headerLocation, uploadURL)
	}), nil
}

// GetBlobUploadStatus 查询上传状态。
func (c *Controller) GetBlobUploadStatus(ctx context.Context, uploadReq *ContainerUploadRequest) (HttpResponseWriter, error) {
	// TODO: 权限检查
	return c.getUploadState(ctx, uploadReq.ContainerReq, uploadReq.Resume)
}

// AppendBlobContent 追加 blob 内容（PATCH）。
func (c *Controller) AppendBlobContent(ctx context.Context, r *http.Request, uploadReq *ContainerUploadRequest) (HttpResponseWriter, error) {
	// TODO: 权限检查
	log := logger.GetLogger(ctx)
	resumeReq := uploadReq.Resume
	if err := resumeReq.AppendWrite(ctx, r); err != nil {
		log.Errorf("append blob content failed: %v", err)
		return nil, err
	}
	log.Infof("append blob content successfully")
	return c.getUploadState(ctx, uploadReq.ContainerReq, resumeReq)
}

// FinishBlobUpload 完成 blob 上传（PUT）。
func (c *Controller) FinishBlobUpload(ctx context.Context, r *http.Request, uploadReq *ContainerUploadRequest) (HttpResponseWriter, error) {
	// TODO: 权限检查
	dgstStr := r.FormValue("digest")
	if dgstStr == "" {
		return nil, ctnAdapter.ErrDigestInvalid.WithDetail("digest missing")
	}
	dgst, err := digest.Parse(dgstStr)
	if err != nil {
		return nil, ctnAdapter.ErrDigestInvalid.WithDetail("digest parsing failed")
	}

	resumeReq := uploadReq.Resume
	if err = resumeReq.Finish(ctx, r, &dgst); err != nil {
		return nil, err
	}

	return NewResponseWriter(func(w http.ResponseWriter) {
		uploadURL := fmt.Sprintf("/v2/%s/blobs/uploads/%s", uploadReq.FullName(), dgst.String())
		w.Header().Add(headerLocation, uploadURL)
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerDigest, dgst.String())
	}), nil
}

// CancelBlobUpload 取消上传会话。
func (c *Controller) CancelBlobUpload(ctx context.Context, uploadReq *ContainerUploadRequest) (HttpResponseWriter, error) {
	// TODO: 权限检查
	resumeReq := uploadReq.Resume
	if err := resumeReq.Remove(ctx); err != nil {
		return nil, err
	}
	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(HeadDockerUUID, resumeReq.SessionID)
		w.WriteHeader(http.StatusNoContent)
	}), nil
}

// ---- Manifest 操作 ----

// HeadContainerManifest HEAD manifest by digest。
func (c *Controller) HeadContainerManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerManifest(ctx, req, manifestReq, true)
}

// GetContainerManifest GET manifest by digest。
func (c *Controller) GetContainerManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerManifest(ctx, req, manifestReq, false)
}

func (c *Controller) getContainerManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest, head bool) (HttpResponseWriter, error) {
	// TODO: 权限检查
	meta, err := c.GetAssetInfo(ctx, manifestReq.Digest.String(), model.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}
	for _, headerVal := range req.Header["If-None-Match"] {
		if headerVal == meta.Path || headerVal == fmt.Sprintf(`"%s"`, meta.Path) {
			return NewResponseWriter(func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotModified)
			}), nil
		}
	}
	return c.getManifestWriter(ctx, manifestReq.View, meta, !head)
}

// PutManifest PUT manifest by digest。
func (c *Controller) PutManifest(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	// TODO: 权限检查
	refDigest := manifestReq.Digest.String()
	uploader := ctnAdapter.NewManifestUploader(c.artStore, manifestReq.View, refDigest)
	if err := handleUpload(ctx, req, uploader); err != nil {
		return nil, err
	}
	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerDigest, refDigest)
		w.WriteHeader(http.StatusCreated)
	}), nil
}

// DeleteContainerManifest DELETE manifest by digest。
func (c *Controller) DeleteContainerManifest(ctx context.Context, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.DeleteContainerBlob(ctx, manifestReq.ContainerReq, manifestReq.Digest.String())
}

func (c *Controller) getManifestWriter(ctx context.Context, view *adapter.ViewDescriptor, desc *AssetMeta, sendContent bool) (HttpResponseWriter, error) {
	content, err := view.Store.Get(ctx, adapter.BlobPath(desc.Ref))
	if err != nil {
		return nil, err
	}
	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerEtag, fmt.Sprintf(`"%s"`, desc.Path))
		w.Header().Add(headerContentType, desc.ContentType)
		w.Header().Add(headerContentLength, fmt.Sprintf("%d", desc.Size))
		w.Header().Add(headerDigest, desc.Path)
		w.WriteHeader(http.StatusOK)
		if sendContent {
			w.Write(content)
		}
	}), nil
}

// ---- Tag 操作 ----

// HeadContainerTag HEAD tag。
func (c *Controller) HeadContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerTag(ctx, req, manifestReq, true)
}

// GetContainerTag GET tag。
func (c *Controller) GetContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	return c.getContainerTag(ctx, req, manifestReq, false)
}

func (c *Controller) getContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest, head bool) (HttpResponseWriter, error) {
	// TODO: 权限检查
	view := manifestReq.View
	t, err := findContainerTag(ctx, c.artStore, view.ViewID, manifestReq.RepoName, manifestReq.GroupName, manifestReq.Tag)
	if err != nil {
		return nil, err
	}
	meta := newAssetMeta(t.asset, t.blob)
	for _, headerVal := range req.Header["If-None-Match"] {
		if headerVal == meta.Path || headerVal == fmt.Sprintf(`"%s"`, meta.Path) {
			return NewResponseWriter(func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotModified)
			}), nil
		}
	}
	return c.getManifestWriter(ctx, view, meta, !head)
}

// PutContainerTag PUT tag（manifest 上传到 tag）。
func (c *Controller) PutContainerTag(ctx context.Context, req *http.Request, manifestReq *ContainerManifestRequest) (HttpResponseWriter, error) {
	// TODO: 权限检查
	view := manifestReq.View

	// 有 GroupName 时自动创建 group，并将 groupId 与 namespace 传入 descriptor
	groupId, err := c.getOrCreateGroup(ctx, view.ViewID, manifestReq.GroupName)
	if err != nil {
		return nil, err
	}

	uploader := ctnAdapter.NewManifestUploader(c.artStore, view, "")
	uploader.Prepare(func(desc *adapter.PackageDescriptor) {
		desc.Name      = manifestReq.RepoName
		desc.Namespace = manifestReq.GroupName // 作为包的命名空间，区分不同 group 下的同名镜像
		desc.GroupId   = groupId               // 供树节点挂载使用
		desc.Version   = manifestReq.Tag
	})
	if err = handleUpload(ctx, req, uploader); err != nil {
		return nil, err
	}
	desc := uploader.Descriptor()
	return NewResponseWriter(func(w http.ResponseWriter) {
		w.Header().Add(headerContentLength, "0")
		w.Header().Add(headerDigest, desc.MainAsset.Path)
		w.WriteHeader(http.StatusCreated)
	}), nil
}

// getOrCreateGroup 在指定 view 下按 groupName 查找或创建 container 格式的一级 group。
// 使用 RecursiveCreate 保证幂等性（group 存在则直接返回）。
func (c *Controller) getOrCreateGroup(ctx context.Context, viewID int64, groupName string) (int64, error) {
	ge, err := c.artStore.Groups().RecursiveCreate(ctx, viewID, []string{groupName}, model.ArtifactContainerFormat)
	if err != nil {
		return 0, err
	}
	return ge.ID, nil
}

// DeleteContainerTag DELETE tag。
func (c *Controller) DeleteContainerTag(ctx context.Context, req *ContainerManifestRequest) (HttpResponseWriter, error) {
	// TODO: 权限检查
	view := req.View
	t, err := findContainerTag(ctx, c.artStore, view.ViewID, req.RepoName, req.GroupName, req.Tag)
	if err != nil {
		return nil, err
	}
	if err = c.artStore.Versions().DeleteById(ctx, t.version.ID); err != nil {
		return nil, err
	}
	if err = c.artStore.Assets().DeleteById(ctx, t.asset.ID); err != nil {
		return nil, err
	}
	if err = c.artStore.Blobs().DeleteById(ctx, t.blob.ID); err != nil {
		return nil, err
	}
	_ = c.fileStore.Delete(ctx, adapter.BlobPath(t.blob.Ref))

	return NewResponseWriter(func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusAccepted)
	}), nil
}

// ---- Tag List ----

type tagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// ListTags GET /v2/{scope}/{viewname}/{groupname}/{name}/tags/list
func (c *Controller) ListTags(ctx context.Context, req *ContainerReq) (*tagList, error) {
	// TODO: 权限检查
	var result = tagList{
		Name: req.RepoName,
		Tags: make([]string, 0),
	}

	pkgModel, err := c.artStore.Packages().GetByName(ctx, req.RepoName, req.GroupName, req.View.ViewID, model.ArtifactContainerFormat)
	if err != nil {
		return &result, nil
	}

	objects, err := c.artStore.Versions().Find(ctx, model.SearchVersionOption{
		PackageId: pkgModel.ID, ViewId: req.View.ViewID,
	})
	if err != nil {
		return &result, nil
	}
	for _, object := range objects {
		result.Tags = append(result.Tags, object.Version)
	}
	return &result, nil
}

// ---- 工具函数 ----

// renderReader 将 reader 内容写入 HTTP response（流式）。
func renderReader(_ context.Context, w http.ResponseWriter, statusCode int, reader io.Reader) {
	w.WriteHeader(statusCode)
	io.Copy(w, reader)
}
