package generic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/request"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
)

const (
	// maxGenericGroupSegmentLen 单段分组名最大 rune 数，与制品库视图名保持一致的上限。
	maxGenericGroupSegmentLen = 128
)

// genericGroupForbiddenChars generic 分组名禁用的字符集合。
// 与视图名黑名单对齐：路径分隔、控制转义、Windows 保留字符；
// `.` 作为层级分隔符由 split 处理，不列入此处（否则每段都会命中）。
var genericGroupForbiddenChars = []rune{'\\', ':', '*', '?', '"', '<', '>', '|'}

type uploader struct {
	uploadReq   *request.ArtifactUploadRequest
	store       storage.ContentStorage
	artStore    artifactStore.Interface
	view        *adapter.ViewDescriptor
	storeLayout adapter.StorageLayout
	descriptor  *adapter.PackageDescriptor
	bufReader   io.Reader
}

// NewUploader 创建 generic 格式的上传器。
func NewUploader(contentStore storage.ContentStorage, artStore artifactStore.Interface, view *adapter.ViewDescriptor) adapter.ArtifactPackageUploader {
	return &uploader{
		uploadReq:   request.NewUpload(view, artStore),
		store:       contentStore,
		artStore:    artStore,
		view:        view,
		storeLayout: adapter.StorageLayoutBlob,
		descriptor:  adapter.NewEmptyPackageDescriptor(),
	}
}

func (h *uploader) Serve(ctx context.Context, req *http.Request) (int64, error) {
	_ = req.ParseMultipartForm(32 << 20)
	if err := h.validateForm(ctx, req); err != nil {
		return 0, err
	}

	file, fh, err := req.FormFile("file")
	if err != nil {
		return 0, err
	}

	fw, ref, err := adapter.NewRandomBlobWriter(ctx, h.store)
	if err != nil {
		return 0, err
	}

	h.uploadReq.RegisterWriter(fw)
	size, hash, err := request.Write(file, fw)

	// 文件名以 form 字段 `name` 为准（由 controller 层的 SpaceRepoFormReq.Validate 保证非空）。
	// 若出于兼容 / 兜底需要 `name` 为空，退回 multipart 的 fh.Filename 以免 asset.path 落空。
	fileName := strings.TrimSpace(req.FormValue("name"))
	if fileName == "" {
		fileName = fh.Filename
	}
	h.descriptor.MainAsset.Path = fileName
	h.descriptor.MainAsset.Size = size
	h.descriptor.MainAsset.Hash = hash
	h.descriptor.MainAsset.Ref = ref
	h.descriptor.MainAsset.ContentType = "application/octet-stream"
	h.descriptor.MainAsset.Kind = model.AssetKindMain
	h.descriptor.MainAsset.Attr = adapter.AttrAssetNormal
	h.descriptor.MainAsset.Format = model.ArtifactFileFormat

	h.descriptor.Format = model.ArtifactFileFormat
	return size, err
}

func (h *uploader) validateForm(_ context.Context, req *http.Request) error {
	group := req.FormValue("group")
	groupID := req.FormValue("groupID")

	if group == "" && groupID == "" {
		return badRequest(
			fmt.Errorf("%w: required parameter 'group' or 'groupID'", adapter.ErrMissFormField),
			i18n.ArtifactUploadGroupRequired,
		)
	}

	artifactID := h.view.ArtifactID
	if groupID != "" {
		gId, err := strconv.ParseInt(groupID, 10, 64)
		if err != nil {
			return badRequest(
				fmt.Errorf("%w: %s", adapter.ErrInvalidGroupId, groupID),
				i18n.ArtifactUploadGroupIdInvalid,
			)
		}
		g, err := h.artStore.Groups().GetById(req.Context(), gId)
		if err != nil {
			return badRequest(
				fmt.Errorf("%w: %s: %v", adapter.ErrInvalidGroupId, groupID, err),
				i18n.ArtifactUploadGroupIdInvalid,
			)
		}
		if g.ArtifactID != artifactID {
			return badRequest(
				fmt.Errorf("%w: groupID=%s does not belong to artifactID=%d", adapter.ErrInvalidGroupId, groupID, artifactID),
				i18n.ArtifactUploadGroupArtifactMismatch,
			)
		}
		h.descriptor.GroupID = gId
	} else {
		if err := validateGenericGroupName(group); err != nil {
			return err
		}
		g, err := h.artStore.Groups().GetEndpointByNames(req.Context(), artifactID, strings.Split(strings.Trim(strings.ReplaceAll(group, "/", "."), "."), "."), model.ArtifactFileFormat)
		if err != nil {
			return badRequest(
				fmt.Errorf("%w: group=%q not found under artifactID=%d: %v", adapter.ErrInvalidGroupName, group, artifactID, err),
				i18n.ArtifactUploadGroupPathNotFound,
			)
		}
		if g.ArtifactGroup == nil || g.ArtifactGroup.ArtifactID != artifactID {
			return badRequest(
				fmt.Errorf("%w: group=%q does not belong to artifactID=%d", adapter.ErrInvalidGroupName, group, artifactID),
				i18n.ArtifactUploadGroupArtifactMismatch,
			)
		}
		h.descriptor.GroupID = g.ID
	}

	return nil
}

// validateGenericGroupName 校验 generic 上传时传入的 group 字符串。
// 约束（与 CreateArtifactViewReq 对 generic/raw 放行中文的口径一致）：
//  1. 以 `.` 拆分后每一段都不能为空（不允许前导 / 尾随 / 连续 `.`）；
//  2. 每段 rune 数不得超过 maxGenericGroupSegmentLen；
//  3. 禁止空白字符（含全角空格等），禁止控制字符；
//  4. 禁止路径分隔与 Windows 保留字符：`/ \ : * ? " < > |`；
//  5. 允许中文、字母、数字、中划线、下划线及其他非保留 Unicode。
func validateGenericGroupName(group string) error {
	if group == "" {
		return badRequest(adapter.ErrInvalidGroupName, i18n.ArtifactUploadGroupNameInvalid)
	}
	parts := strings.Split(group, ".")
	for _, seg := range parts {
		if seg == "" {
			return badRequest(
				fmt.Errorf("%w: empty segment in group=%q", adapter.ErrInvalidGroupName, group),
				i18n.ArtifactUploadGroupNameInvalid,
			)
		}
		if len([]rune(seg)) > maxGenericGroupSegmentLen {
			return badRequest(
				fmt.Errorf("%w: segment too long in group=%q", adapter.ErrInvalidGroupName, group),
				i18n.ArtifactUploadGroupNameInvalid,
			)
		}
		for _, rn := range seg {
			if unicode.IsSpace(rn) || unicode.IsControl(rn) {
				return badRequest(
					fmt.Errorf("%w: group segment contains whitespace or control char: %q", adapter.ErrInvalidGroupName, seg),
					i18n.ArtifactUploadGroupNameInvalid,
				)
			}
			for _, bad := range genericGroupForbiddenChars {
				if rn == bad {
					return badRequest(
						fmt.Errorf("%w: group segment contains forbidden char %q: %q", adapter.ErrInvalidGroupName, bad, seg),
						i18n.ArtifactUploadGroupNameInvalid,
					)
				}
			}
		}
	}
	return nil
}

// badRequest 将错误标记为 400 状态并附带 i18n key，供 controller 层据此选择响应状态码。
func badRequest(err error, key i18n.I18nKey) error {
	return pkgErrors.WithHttpStatus(pkgErrors.WithI18nKey(err, key), http.StatusBadRequest)
}

func (h *uploader) IsValid(ctx context.Context) error {
	u := h.uploadReq.LoadCreator(ctx)
	h.descriptor.VersionMetadata = &adapter.VersionMetadata{CreatorName: u.UID}
	h.uploadReq.Descriptor = h.descriptor
	return nil
}

func (h *uploader) Save(ctx context.Context) error {
	if err := h.uploadReq.Commit(ctx); err != nil {
		_ = h.uploadReq.Cancel(ctx)
		return err
	}
	return nil
}

func (h *uploader) Cancel(ctx context.Context) error {
	return h.uploadReq.Cancel(ctx)
}

func (h *uploader) GetAssets() []*adapter.SavedAsset {
	return h.uploadReq.Assets
}
