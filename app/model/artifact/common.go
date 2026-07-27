package artifact

import (
	"fmt"
	"strings"
	"unicode"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	common "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/common"
)

// maxArtifactViewNameLen 制品库视图名称的最大长度（与 DB 字段保持一致的冗余保护）。
const maxArtifactViewNameLen = 128

func ValidateLibName(name string, format ArtifactFormat) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.WithI18nKey(
			fmt.Errorf("artifact lib name is required"),
			i18n.ArtifactLibNameRequired,
		)
	}
	if len([]rune(name)) > maxArtifactViewNameLen {
		return "", errors.WithI18nKey(
			fmt.Errorf("artifact lib name too long: %d runes", len([]rune(name))),
			i18n.ArtifactLibNameTooLong,
		)
	}

	var valid bool
	if format == ArtifactFileFormat {
		valid = !strings.ContainsFunc(name, unicode.IsSpace) &&
			!strings.ContainsAny(name, "\\/\"")
	} else {
		valid = !strings.ContainsFunc(name, unicode.IsSpace) && common.IsValidCode(name)
	}

	if !valid {
		return "", errors.WithI18nKey(
			fmt.Errorf("artifact lib name contains invalid characters: %q", name),
			i18n.ArtifactLibNameInvalidChars,
		)
	}

	return name, nil
}

// SpaceRef 是给 artifact adapter 使用的精简空间引用。
type SpaceRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// RepoRef 是给 artifact adapter 使用的精简仓库引用。
type RepoRef struct {
	ID     int64  `json:"id"`
	GitUID string `json:"gitUID"`
	Name   string `json:"name"`
}

// CreatorInfo 保存 artifact 操作发起人的基础身份信息。
// CreatorInfo holds the creator information for artifact operations.
type CreatorInfo struct {
	ID          int64  `json:"id"`
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
}

// SpaceRepoReq 定义非文件上传场景使用的 JSON 请求参数。
type SpaceRepoReq struct {
	SpaceID    int64 `json:"spaceID"`
	RepoID     int64 `json:"repoID"`
	ArtifactID int64 `json:"artifactID"`
}

// SpaceRepoFormReq 定义文件上传场景使用的 Form 请求参数。
// 上传必须通过 artifactID 精确定位制品库视图，禁止再通过 space/repo 回落推导。
// 注意：此处不在 tag 层使用 binding:required，以便在控制器里抛出更友好的多语言错误。
type SpaceRepoFormReq struct {
	ArtifactID  int64  `form:"artifactID"`
	Name        string `form:"name"`
	PackageName string `form:"packageName"`
}

func (r *SpaceRepoFormReq) Validate() error {
	if r.ArtifactID <= 0 {
		return errors.WithI18nKey(
			fmt.Errorf("artifactID is required and must be greater than 0, got %d", r.ArtifactID),
			i18n.ArtifactUploadArtifactidRequired,
		)
	}

	// 文件名必填：generic/raw 用它覆盖 asset.path。
	// 做 TrimSpace 后再判空，避免纯空白被当作有效值。
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return errors.WithI18nKey(
			fmt.Errorf("upload form field 'name' is required"),
			i18n.ArtifactUploadFileNameRequired,
		)
	}

	// packageName 是 raw 的包名字段，由 raw adapter 负责做格式级必填和正则校验。
	r.PackageName = strings.TrimSpace(r.PackageName)
	return nil
}

type SoftRemoveEntitiesReq struct {
	EntityIDs []EntityID `json:"entityIDs,omitempty"`
}

// ListNodeInfoReq 兼容参考项目 `/artifacts/nodeInfo` 的请求体。
type ListNodeInfoReq struct {
	EntityIDs []string `json:"entityIDs,omitempty"`
}

type CreateArtifactLibReq struct {
	SpaceRepoReq
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Format      string `json:"format"`
}

// Validate 对 CreateArtifactLibReq 进行业务规则校验
func (r *CreateArtifactLibReq) Validate() error {
	if strings.TrimSpace(r.Format) == "" {
		return errors.WithI18nKey(
			fmt.Errorf("artifact format is required"),
			i18n.ArtifactLibFormatRequired,
		)
	}

	// 2. Format 枚举校验（使用 constant 包的 Sanitize 模式）
	format := enum.ArtifactFormat(r.Format)
	if _, ok := format.Sanitize(); !ok {
		return errors.WithI18nKey(
			fmt.Errorf("unsupported format: %s", r.Format),
			i18n.ArtifactLibFormatInvalid,
		)
	}

	// 3. Name 合法性校验（复用 ValidateViewName）
	name, err := ValidateLibName(r.Name, ArtifactFormat(r.Format))
	if err != nil {
		return err
	}
	r.Name = name

	if !common.IsValidCode(r.Code) {
		return errors.WithI18nKey(nil, i18n.ArtifactLibCodeInvalid)
	}

	return nil
}

// RefreshTreeReq 接收 `/artifacts/tree/refresh` 的请求参数。
// Format 必填：值必须通过 enum.ArtifactFormat.Sanitize() 校验，或为 ArtifactAllFormat ("all")
// 以表示全量刷新所有格式。
type RefreshTreeReq struct {
	Format string `json:"format"`
}

// Validate 校验 RefreshTreeReq。
func (r *RefreshTreeReq) Validate() error {
	f := strings.TrimSpace(r.Format)
	if f == "" {
		f = string(ArtifactAllFormat)
	}
	// 允许 "all" 表示全量刷新
	if ArtifactFormat(f) == ArtifactAllFormat {
		r.Format = f
		return nil
	}
	if _, ok := enum.ArtifactFormat(f).Sanitize(); !ok {
		return errors.WithI18nKey(
			fmt.Errorf("unsupported format: %s", r.Format),
			i18n.ArtifactTreeRefreshFormatInvalid,
		)
	}
	r.Format = f
	return nil
}

// ArtifactVersionListReq 接收获取版本列表的请求参数。
type ArtifactVersionListReq struct {
	SpaceRepoReq
	PackageName string `json:"packageName"`
}

// Validate 对版本列表请求做业务校验，返回携带 i18n key 的错误。
func (r *ArtifactVersionListReq) Validate() error {
	if r.ArtifactID <= 0 {
		return errors.WithI18nKey(
			fmt.Errorf("artifactID is required and must be greater than 0, got %d", r.ArtifactID),
			i18n.ArtifactVersionArtifactidRequired,
		)
	}
	r.PackageName = strings.TrimSpace(r.PackageName)
	if r.PackageName == "" {
		return errors.WithI18nKey(
			fmt.Errorf("packageName is required"),
			i18n.ArtifactVersionPackageRequired,
		)
	}
	return nil
}

type ArtifactTreeListReq struct {
	SpaceRepoReq
	ArtifactTreeFilter
}

// ListAssetReq 制品资产分页列表请求。
// EntityID 接受 "group.{id}" / "version.{id}" / "package.{id}" / "asset.{id}"。
type ListAssetReq struct {
	EntityID  EntityID `json:"entityID" binding:"required"` // group.{id} / version.{id} / package.{id} / asset.{id}
	Page      int      `json:"page"`
	PageSize  int      `json:"pageSize"`
	Name      string   `json:"name"`      // 按文件名模糊过滤
	Version   string   `json:"version"`   // 按版本号过滤（仅 version 类资产有效）
	CreatedBy string   `json:"createdBy"` // 按创建人过滤
	EditedBy  string   `json:"editedBy"`  // 按更新人过滤
	Sort      string   `json:"sort"`      // 排序字段：id / name / size / createdDate / editedDate
	Order     string   `json:"order"`     // 排序方向：asc / desc
	Deleted   *bool    `json:"deleted"`   // 是否包含已删除：nil=不过滤，true=仅已删，false=仅未删
	More      bool     `json:"more"`      // 是否展开子目录制品放在同一列表中
}

// CreateGroupReq 创建 group 节点的请求参数。
type CreateGroupReq struct {
	ArtifactID int64  `json:"artifactID"`
	Names      string `json:"names" binding:"required"`
	Format     string `json:"format"`
}

// ParseNames 将 Names 按 "/" 拆分为分组路径段。
func (r *CreateGroupReq) ParseNames() []string {
	return strings.Split(r.Names, "/")
}

// Validate 对 CreateGroupReq 进行业务规则校验
func (r *CreateGroupReq) Validate() error {
	// 0. 规范化分隔符：统一使用 "/" 作为路径分隔符，去除首尾 "/"
	r.Names = strings.Trim(r.Names, "/")

	// 1. ArtifactID 校验
	if r.ArtifactID <= 0 {
		return errors.WithI18nKey(
			fmt.Errorf("artifactID must be greater than 0, got %d", r.ArtifactID),
			i18n.ArtifactGroupArtifactidRequired,
		)
	}

	// 3. Names 合法性校验：先拒空段，再校验每段字符集与首尾空白。
	parts := strings.Split(r.Names, "/")
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return errors.WithI18nKey(
				fmt.Errorf("group names contain empty segment: %q", r.Names),
				i18n.ArtifactGroupNamesInvalid,
			)
		}
		if err := validateGroupSegment(part); err != nil {
			return err
		}
	}

	// 4. container 格式不允许嵌套
	if enum.ArtifactFormat(r.Format) == enum.ArtifactFormatContainer && len(parts) > 1 {
		return errors.WithI18nKey(
			fmt.Errorf("container format does not support nested groups: %q", r.Names),
			i18n.ArtifactGroupContainerNoNesting,
		)
	}

	return nil
}

// groupNameForbiddenChars 任何 format 下都禁止在 group 单段名称中出现的字符。
var groupNameForbiddenChars = []rune{'\\', ':', '*', '?', '"', '<', '>', '|'}

// validateGroupSegment 校验 group 单段名称的合法性。
func validateGroupSegment(seg string) error {
	if seg == "" {
		return errors.WithI18nKey(
			fmt.Errorf("group name segment is empty"),
			i18n.ArtifactGroupNamesInvalid,
		)
	}
	runes := []rune(seg)
	if unicode.IsSpace(runes[0]) || unicode.IsSpace(runes[len(runes)-1]) {
		return errors.WithI18nKey(
			fmt.Errorf("group name segment must not start or end with whitespace: %q", seg),
			i18n.ArtifactGroupNameLeadingTrailingSpace,
		)
	}
	for _, rn := range runes {
		if unicode.IsSpace(rn) {
			return errors.WithI18nKey(
				fmt.Errorf("group name segment contains whitespace: %q", seg),
				i18n.ArtifactGroupNameInvalidChars,
			)
		}
		for _, bad := range groupNameForbiddenChars {
			if rn == bad {
				return errors.WithI18nKey(
					fmt.Errorf("group name segment contains forbidden char %q: %q", bad, seg),
					i18n.ArtifactGroupNameInvalidChars,
				)
			}
		}
	}
	return nil
}
