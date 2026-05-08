package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	ctnAdapter "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	ctnCtrl "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/controller/container"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	spaceStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"
)

// ContainerHandler 容器镜像 Gin Handler。
// 将 OCI v2 协议请求适配到 ContainerController。
type ContainerHandler struct {
	ctrl       *ctnCtrl.Controller
	artSvc     artifactSvc.Interface
	artStore   artifactStore.Interface
	spaceStore spaceStore.Interface
}

// NewContainerHandler 创建容器镜像 Handler 实例。
func NewContainerHandler(
	ctrl *ctnCtrl.Controller,
	artSvc artifactSvc.Interface,
	artStore artifactStore.Interface,
	spStore spaceStore.Interface,
) *ContainerHandler {
	return &ContainerHandler{
		ctrl:       ctrl,
		artSvc:     artSvc,
		artStore:   artStore,
		spaceStore: spStore,
	}
}

// HandleAPIBase OCI v2 协议探测
// @Summary OCI v2 协议探测
// @Description Docker Registry API v2 探测端点，用于检测服务是否支持 OCI 分发规范。docker login 时自动调用。
// @Tags container
// @Produce application/json
// @Success 200 {object} map[string]interface{} "空 JSON 对象，表示支持 v2 协议"
// @Failure 401 {object} map[string]interface{} "未认证，响应包含 WWW-Authenticate 挑战头"
// @Router /v2/ [get]
func (h *ContainerHandler) HandleAPIBase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// HandleBlobHead 检查 Blob 是否存在
// @Summary 检查 Blob 是否存在
// @Description 根据 digest 检查指定容器镜像空间下的 Blob 是否存在，返回 Content-Length 等元信息。
// @Tags container
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param digest path string true "Blob 摘要 (如 sha256:abc123...)"
// @Success 200 "Blob 存在，响应头包含 Content-Length 和 Docker-Content-Digest"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "Blob 不存在"
// @Router /v2/{space}/{name}/blobs/{digest} [head]
func (h *ContainerHandler) HandleBlobHead(c *gin.Context) {
	ctx := c.Request.Context()
	ctnReq, err := h.loadContainerRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	dgst := c.Param("digest")
	if _, e := digest.Parse(dgst); e != nil {
		renderContainerErr(c, ctnAdapter.ErrDigestInvalid)
		return
	}
	rw, err := h.ctrl.HeadBlob(ctx, ctnReq, dgst)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
	c.Writer.WriteHeader(http.StatusOK)
}

// HandleBlobGet 下载 Blob
// @Summary 下载 Blob
// @Description 根据 digest 下载指定容器镜像空间下的 Blob 内容（镜像层数据）。
// @Tags container
// @Produce application/octet-stream
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param digest path string true "Blob 摘要 (如 sha256:abc123...)"
// @Success 200 {file} binary "Blob 内容"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "Blob 不存在"
// @Router /v2/{space}/{name}/blobs/{digest} [get]
func (h *ContainerHandler) HandleBlobGet(c *gin.Context) {
	ctx := c.Request.Context()
	ctnReq, err := h.loadContainerRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	dgst := c.Param("digest")
	if _, e := digest.Parse(dgst); e != nil {
		renderContainerErr(c, ctnAdapter.ErrDigestInvalid)
		return
	}
	rw, err := h.ctrl.GetContainerBlob(ctx, ctnReq, dgst)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
}

// HandleBlobDelete 删除 Blob
// @Summary 删除 Blob
// @Description 根据 digest 删除指定容器镜像空间下的 Blob。
// @Tags container
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param digest path string true "Blob 摘要 (如 sha256:abc123...)"
// @Success 202 "删除成功"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "Blob 不存在"
// @Router /v2/{space}/{name}/blobs/{digest} [delete]
func (h *ContainerHandler) HandleBlobDelete(c *gin.Context) {
	ctx := c.Request.Context()
	ctnReq, err := h.loadContainerRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	dgst := c.Param("digest")
	rw, err := h.ctrl.DeleteContainerBlob(ctx, ctnReq, dgst)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
}

// HandleBlobUploadStart 发起 Blob 上传
// @Summary 发起 Blob 上传
// @Description 发起一个新的 Blob 上传会话。如果 query 中包含 digest 参数则执行单次上传（monolithic），否则初始化分块上传。
// @Tags container
// @Accept application/octet-stream
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param digest query string false "Blob 摘要（传入时执行单次上传）"
// @Success 202 "上传会话已创建，响应头包含 Location 和 Docker-Upload-UUID"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "空间或镜像不存在"
// @Router /v2/{space}/{name}/blobs/uploads/ [post]
func (h *ContainerHandler) HandleBlobUploadStart(c *gin.Context) {
	ctx := c.Request.Context()
	// cross mount 不支持
	if c.Query("from") != "" || c.Query("mount") != "" {
		renderContainerErr(c, ctnAdapter.ErrUnsupported.WithDetail("cross mount is not supported"))
		return
	}

	ctnReq, err := h.loadContainerRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}

	var rw ctnCtrl.HttpResponseWriter
	paramDigest := c.Query("digest")
	if paramDigest != "" {
		dgst, e := digest.Parse(paramDigest)
		if e != nil {
			renderContainerErr(c, ctnAdapter.ErrDigestInvalid)
			return
		}
		rw, err = h.ctrl.SaveContainerBlob(ctx, ctnReq, c.Request, dgst)
	} else {
		rw, err = h.ctrl.StartBlobUpload(ctx, ctnReq)
	}

	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
	c.Writer.WriteHeader(http.StatusAccepted)
}

// HandleBlobUploadPatch 追加 Blob 分块数据
// @Summary 追加 Blob 分块数据
// @Description 向正在进行的分块上传会话追加数据。请求体为原始字节流。
// @Tags container
// @Accept application/octet-stream
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param uuid path string true "上传会话 UUID"
// @Param _state query string true "上传状态 token"
// @Success 202 "数据追加成功，响应头包含 Range 和 Docker-Upload-UUID"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "上传会话不存在"
// @Router /v2/{space}/{name}/blobs/uploads/{uuid} [patch]
func (h *ContainerHandler) HandleBlobUploadPatch(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.GetLoggerFromGin(c)
	uploadReq, err := h.loadUploadRequest(c, true)
	if err != nil {
		log.Errorf("【container push】PATCH load upload failed: path=%s err=%v", c.Request.URL.Path, err)
		renderContainerErr(c, err)
		return
	}
	rw, err := h.ctrl.AppendBlobContent(ctx, c.Request, uploadReq)
	if err != nil {
		log.Errorf("【container push】PATCH append failed: session=%s err=%v", uploadReq.Resume.SessionID, err)
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
	c.Writer.WriteHeader(http.StatusAccepted)
}

// HandleBlobUploadFinish 完成 Blob 上传
// @Summary 完成 Blob 上传
// @Description 完成分块上传会话，可选附带最后一块数据。必须在 query 中提供最终 digest 校验。
// @Tags container
// @Accept application/octet-stream
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param uuid path string true "上传会话 UUID"
// @Param digest query string true "最终 Blob 摘要"
// @Param _state query string true "上传状态 token"
// @Success 201 "上传完成，响应头包含 Docker-Content-Digest 和 Location"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "上传会话不存在"
// @Router /v2/{space}/{name}/blobs/uploads/{uuid} [put]
func (h *ContainerHandler) HandleBlobUploadFinish(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.GetLoggerFromGin(c)
	uploadReq, err := h.loadUploadRequest(c, true)
	if err != nil {
		log.Errorf("【container push】PUT load upload failed: path=%s err=%v", c.Request.URL.Path, err)
		renderContainerErr(c, err)
		return
	}
	rw, err := h.ctrl.FinishBlobUpload(ctx, c.Request, uploadReq)
	if err != nil {
		log.Errorf("【container push】PUT finish failed: session=%s err=%v", uploadReq.Resume.SessionID, err)
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
	c.Writer.WriteHeader(http.StatusCreated)
}

// HandleBlobUploadCancel 取消 Blob 上传
// @Summary 取消 Blob 上传
// @Description 取消正在进行的分块上传会话，释放已上传的临时数据。
// @Tags container
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param uuid path string true "上传会话 UUID"
// @Param _state query string true "上传状态 token"
// @Success 204 "取消成功"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "上传会话不存在"
// @Router /v2/{space}/{name}/blobs/uploads/{uuid} [delete]
func (h *ContainerHandler) HandleBlobUploadCancel(c *gin.Context) {
	ctx := c.Request.Context()
	uploadReq, err := h.loadUploadRequest(c, true)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw, err := h.ctrl.CancelBlobUpload(ctx, uploadReq)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
	c.Writer.WriteHeader(http.StatusNoContent)
}

// HandleBlobUploadStatus 查询 Blob 上传状态
// @Summary 查询 Blob 上传状态
// @Description 查询分块上传会话的当前状态，返回已上传的字节范围。
// @Tags container
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param uuid path string true "上传会话 UUID"
// @Success 204 "查询成功，响应头包含 Range 和 Docker-Upload-UUID"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "上传会话不存在"
// @Router /v2/{space}/{name}/blobs/uploads/{uuid} [get]
func (h *ContainerHandler) HandleBlobUploadStatus(c *gin.Context) {
	ctx := c.Request.Context()
	uploadReq, err := h.loadUploadRequest(c, false)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw, err := h.ctrl.GetBlobUploadStatus(ctx, uploadReq)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
	c.Writer.WriteHeader(http.StatusNoContent)
}

// HandleManifestHead 检查 Manifest 是否存在
// @Summary 检查 Manifest 是否存在
// @Description 根据 reference（tag 或 digest）检查 Manifest 是否存在，返回 Content-Type 和 Docker-Content-Digest。
// @Tags container
// @Produce application/vnd.docker.distribution.manifest.v2+json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param reference path string true "引用（tag 名称或 digest）"
// @Success 200 "Manifest 存在，响应头包含 Content-Length 和 Docker-Content-Digest"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "Manifest 不存在"
// @Router /v2/{space}/{name}/manifests/{reference} [head]
func (h *ContainerHandler) HandleManifestHead(c *gin.Context) {
	ctx := c.Request.Context()
	manifestReq, err := h.loadManifestRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}

	var rw ctnCtrl.HttpResponseWriter
	if manifestReq.IsTag {
		rw, err = h.ctrl.HeadContainerTag(ctx, c.Request, manifestReq)
	} else {
		rw, err = h.ctrl.HeadContainerManifest(ctx, c.Request, manifestReq)
	}
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			renderContainerErr(c, ctnAdapter.ErrManifestUnknown)
		} else {
			renderContainerErr(c, err)
		}
		return
	}
	rw.Write(c.Writer)
}

// HandleManifestGet 下载 Manifest
// @Summary 下载 Manifest
// @Description 根据 reference（tag 或 digest）下载容器镜像的 Manifest 内容。
// @Tags container
// @Produce application/vnd.docker.distribution.manifest.v2+json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param reference path string true "引用（tag 名称或 digest）"
// @Success 200 {object} map[string]interface{} "Manifest JSON 内容"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "Manifest 不存在"
// @Router /v2/{space}/{name}/manifests/{reference} [get]
func (h *ContainerHandler) HandleManifestGet(c *gin.Context) {
	ctx := c.Request.Context()
	manifestReq, err := h.loadManifestRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}

	var rw ctnCtrl.HttpResponseWriter
	if manifestReq.IsTag {
		rw, err = h.ctrl.GetContainerTag(ctx, c.Request, manifestReq)
	} else {
		rw, err = h.ctrl.GetContainerManifest(ctx, c.Request, manifestReq)
	}
	if err != nil {
		if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
			renderContainerErr(c, ctnAdapter.ErrManifestUnknown)
		} else {
			renderContainerErr(c, err)
		}
		return
	}
	rw.Write(c.Writer)
}

// HandleManifestPut 上传 Manifest
// @Summary 上传 Manifest
// @Description 上传容器镜像的 Manifest（docker push 的核心步骤）。reference 可以是 tag 或 digest。
// @Tags container
// @Accept application/vnd.docker.distribution.manifest.v2+json
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param reference path string true "引用（tag 名称或 digest）"
// @Success 201 "上传成功，响应头包含 Docker-Content-Digest 和 Location"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 400 {object} map[string]interface{} "Manifest 格式错误"
// @Router /v2/{space}/{name}/manifests/{reference} [put]
func (h *ContainerHandler) HandleManifestPut(c *gin.Context) {
	ctx := c.Request.Context()
	manifestReq, err := h.loadManifestRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}

	var rw ctnCtrl.HttpResponseWriter
	if manifestReq.IsTag {
		rw, err = h.ctrl.PutContainerTag(ctx, c.Request, manifestReq)
	} else {
		rw, err = h.ctrl.PutManifest(ctx, c.Request, manifestReq)
	}
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
}

// HandleManifestDelete 删除 Manifest
// @Summary 删除 Manifest
// @Description 根据 reference（tag 或 digest）删除容器镜像的 Manifest 及关联数据。
// @Tags container
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Param reference path string true "引用（tag 名称或 digest）"
// @Success 202 "删除成功"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "Manifest 不存在"
// @Router /v2/{space}/{name}/manifests/{reference} [delete]
func (h *ContainerHandler) HandleManifestDelete(c *gin.Context) {
	ctx := c.Request.Context()
	manifestReq, err := h.loadManifestRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}

	var rw ctnCtrl.HttpResponseWriter
	if manifestReq.IsTag {
		rw, err = h.ctrl.DeleteContainerTag(ctx, manifestReq)
	} else {
		rw, err = h.ctrl.DeleteContainerManifest(ctx, manifestReq)
	}
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	rw.Write(c.Writer)
}

// HandleListTag 列出镜像 Tag
// @Summary 列出镜像 Tag
// @Description 列出指定容器镜像的所有 Tag 列表。
// @Tags container
// @Produce application/json
// @Param space path string true "空间名称"
// @Param name path string true "镜像名称"
// @Success 200 {object} map[string]interface{} "Tag 列表，格式 {name: string, tags: []string}"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "镜像不存在"
// @Router /v2/{space}/{name}/tags/list [get]
func (h *ContainerHandler) HandleListTag(c *gin.Context) {
	ctx := c.Request.Context()
	ctnReq, err := h.loadContainerRequest(c)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	data, err := h.ctrl.ListTags(ctx, ctnReq)
	if err != nil {
		renderContainerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// ---- 内部方法 ----

// detectScope 从 gin 路由 FullPath 前缀判断当前请求的作用域。
// 路由前缀固定为 /v2/global/、/v2/space/、/v2/repo/，不需要额外中间件。
func detectScope(c *gin.Context) string {
	p := c.FullPath()
	switch {
	case strings.HasPrefix(p, "/v2/global/"):
		return "global"
	case strings.HasPrefix(p, "/v2/space/"):
		return "space"
	default:
		return "repo"
	}
}

// loadContainerRequest 从 Gin context 解析容器请求参数。
//
// 路由格式：
//
//	/v2/global/{viewname}/{groupname}/{imagename}
//	/v2/space/{spaceName}/{viewname}/{groupname}/{imagename}
//	/v2/repo/{spaceName}/{repoName}/{viewname}/{groupname}/{imagename}
//
// groupname 和 imagename 均作为独立的具名参数绑定，不再使用 catch-all *repopath。
func (h *ContainerHandler) loadContainerRequest(c *gin.Context) (*ctnCtrl.ContainerReq, error) {
	log := logger.GetLoggerFromGin(c)

	scope := detectScope(c)
	viewname := c.Param("viewname")
	spaceName := c.Param("spaceName") // scope=global 时为 ""
	scopeRepo := c.Param("repoName")  // scope=global/space 时为 ""

	groupName := c.Param("groupname")
	imageName := c.Param("imagename")
	if groupName == "" || imageName == "" {
		return nil, ctnAdapter.ErrNameInvalid.WithDetail("groupname and imagename are required")
	}

	ctx := c.Request.Context()
	view, err := h.artSvc.ParseContainerView(ctx, scope, spaceName, scopeRepo, viewname)
	if err != nil {
		log.Errorf("【container】parse view failed: scope=%s space=%s repo=%s view=%s err=%v",
			scope, spaceName, scopeRepo, viewname, err)
		return nil, ctnAdapter.ErrNameUnknown.WithDetail("view not found: " + viewname)
	}

	return &ctnCtrl.ContainerReq{
		Scope:     scope,
		SpaceName: spaceName,
		ScopeRepo: scopeRepo,
		ViewName:  viewname,
		GroupName: groupName,
		RepoName:  imageName,
		View:      view,
	}, nil
}

// loadManifestRequest 解析 manifest 请求（tag 或 digest）。
func (h *ContainerHandler) loadManifestRequest(c *gin.Context) (*ctnCtrl.ContainerManifestRequest, error) {
	baseReq, err := h.loadContainerRequest(c)
	if err != nil {
		return nil, err
	}
	reference := c.Param("reference")
	if reference == "" {
		return nil, ctnAdapter.ErrManifestUnknown.WithDetail("missing reference")
	}
	return ctnCtrl.ParseManifestReference(baseReq, reference)
}

// loadUploadRequest 解析 blob 上传续传请求。
func (h *ContainerHandler) loadUploadRequest(c *gin.Context, requireState bool) (*ctnCtrl.ContainerUploadRequest, error) {
	baseReq, err := h.loadContainerRequest(c)
	if err != nil {
		return nil, err
	}

	paramUUID := c.Param("uuid")
	if paramUUID == "" {
		return nil, errors.New("missing uuid")
	}

	token := c.Query("_state")
	if token == "" && requireState {
		return nil, errors.New("missing _state")
	}

	ctx := c.Request.Context()
	resume, err := ctnAdapter.ParseResumeRequest(ctx, paramUUID, baseReq.FullName(), token, h.artStore, baseReq.View)
	if err != nil {
		log := logger.GetLoggerFromGin(c)
		log.Errorf("【container push】parse resume upload failed: uuid=%s err=%v", paramUUID, err)
		return nil, err
	}

	return &ctnCtrl.ContainerUploadRequest{
		ContainerReq: baseReq,
		Resume:       resume,
	}, nil
}

// renderContainerErr 按 OCI 规范渲染错误（Docker Distribution Spec 固定 JSON 格式）。
func renderContainerErr(c *gin.Context, err error) {
	e := translateErr(err)
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(e.HTTPStatusValue)
	_ = json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"errors": []*ctnAdapter.RegistryErr{e},
	})
	c.Abort()
}

func translateErr(err error) *ctnAdapter.RegistryErr {
	var expect *ctnAdapter.RegistryErr
	if errors.As(err, &expect) {
		return expect
	}
	if pkgErrors.Is(err, pkgErrors.ErrResourceNotFound) {
		return ctnAdapter.ErrBlobUnknown.WithDetail("resource not found")
	}
	return ctnAdapter.ErrUnknown.WithDetail(err.Error())
}
