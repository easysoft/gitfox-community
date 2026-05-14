package artifact

import (
	"fmt"
	"strings"
	"unicode"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

// maxArtifactViewNameLen 制品库视图名称的最大长度（与 DB 字段保持一致的冗余保护）。
const maxArtifactViewNameLen = 128

// viewNameForbiddenChars 任何 format 下都禁止在 view 名称中出现的字符。
// 覆盖路径分隔符、点号（避免与 group 路径分隔号冲突）以及 Windows 保留字符。
var viewNameForbiddenChars = []rune{'/', '\\', '.', ':', '*', '?', '"', '<', '>', '|'}

// ValidateViewName 校验制品库视图名称的合法性。
// 规则：
//  1. 非空，长度 ≤ maxArtifactViewNameLen
//  2. 不得包含空白字符或 viewNameForbiddenChars 中的字符
//  3. 版本化格式（container/helm/pypi/npm/maven）只允许 ASCII 字母、数字、-、_
//
// 返回 trim 后的名称和可能的错误（已带 i18n key）。
func ValidateViewName(name string, format ArtifactFormat) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", errors.WithI18nKey(
			fmt.Errorf("artifact view name is required"),
			i18n.ArtifactViewNameRequired,
		)
	}
	if len([]rune(trimmed)) > maxArtifactViewNameLen {
		return "", errors.WithI18nKey(
			fmt.Errorf("artifact view name too long: %d runes", len([]rune(trimmed))),
			i18n.ArtifactViewNameTooLong,
		)
	}
	for _, rn := range trimmed {
		if unicode.IsSpace(rn) {
			return "", errors.WithI18nKey(
				fmt.Errorf("artifact view name contains whitespace: %q", trimmed),
				i18n.ArtifactViewNameInvalidChars,
			)
		}
		for _, bad := range viewNameForbiddenChars {
			if rn == bad {
				return "", errors.WithI18nKey(
					fmt.Errorf("artifact view name contains forbidden char %q: %q", bad, trimmed),
					i18n.ArtifactViewNameInvalidChars,
				)
			}
		}
	}
	// 版本化格式：只允许 ASCII 字母、数字、中划线、下划线
	if IsVersionedFormat(format) {
		for _, rn := range trimmed {
			if !isVersionedViewNameRune(rn) {
				return "", errors.WithI18nKey(
					fmt.Errorf("versioned artifact view name must be ASCII letters/digits/-/_: %q", trimmed),
					i18n.ArtifactViewNameNonAsciiNotAllowed,
				)
			}
		}
	}
	return trimmed, nil
}

// SpaceRef 是给 artifact adapter 使用的精简空间引用。
type SpaceRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
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
	ArtifactID int64  `form:"artifactID"`
	Name       string `form:"name"`
}

func (r *SpaceRepoFormReq) Validate() error {
	if r.ArtifactID <= 0 {
		return errors.WithI18nKey(
			fmt.Errorf("artifactID is required and must be greater than 0, got %d", r.ArtifactID),
			i18n.ArtifactUploadArtifactidRequired,
		)
	}

	// 文件名必填：generic 用它覆盖 asset.path。
	// 做 TrimSpace 后再判空，避免纯空白被当作有效值。
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return errors.WithI18nKey(
			fmt.Errorf("upload form field 'name' is required"),
			i18n.ArtifactUploadFileNameRequired,
		)
	}

	return nil
}

type SoftRemoveEntitiesReq struct {
	EntityIDs []string `json:"entityIDs,omitempty"`
}

// ListNodeInfoReq 兼容参考项目 `/artifacts/nodeInfo` 的请求体。
type ListNodeInfoReq struct {
	EntityIDs []string `json:"entityIDs,omitempty"`
}

type CreateArtifactViewReq struct {
	SpaceRepoReq
	Name        string `json:"name"`
	Description string `json:"description"`
	Format      string `json:"format"`
}

// Validate 对 CreateArtifactViewReq 进行业务规则校验，返回携带 i18n key 的错误。
// 校验规则：
//  1. Format 必须通过 enum.ArtifactFormat.Sanitize() 校验
//  2. Name 必填且不能超过 maxArtifactViewNameLen
//  3. Name 不得包含空白字符或路径/Windows 保留字符
//  4. 版本化 format（container/helm/pypi/npm/maven）的 Name 仅允许 ASCII 字母、数字、中划线、下划线
//     （generic / raw 允许中文等非 ASCII 字符）
func (r *CreateArtifactViewReq) Validate() error {

	// 1. Format 必填校验：空值给出更友好的提示，避免落到 "不支持的格式"。
	if strings.TrimSpace(r.Format) == "" {
		return errors.WithI18nKey(
			fmt.Errorf("artifact format is required"),
			i18n.ArtifactViewFormatRequired,
		)
	}

	// 2. Format 枚举校验（使用 constant 包的 Sanitize 模式）
	format := enum.ArtifactFormat(r.Format)
	if _, ok := format.Sanitize(); !ok {
		return errors.WithI18nKey(
			fmt.Errorf("unsupported format: %s", r.Format),
			i18n.ArtifactViewFormatInvalid,
		)
	}

	// 3. Name 合法性校验（复用 ValidateViewName）
	name, err := ValidateViewName(r.Name, ArtifactFormat(r.Format))
	if err != nil {
		return err
	}

	// 回写 trim 后的 name，后续 service 直接使用
	r.Name = name

	return nil
}

// isVersionedViewNameRune 报告字符是否可用于版本化制品库名称：ASCII 字母、数字、中划线、下划线。
func isVersionedViewNameRune(rn rune) bool {
	if rn > unicode.MaxASCII {
		return false
	}
	switch {
	case rn >= 'a' && rn <= 'z':
		return true
	case rn >= 'A' && rn <= 'Z':
		return true
	case rn >= '0' && rn <= '9':
		return true
	case rn == '-' || rn == '_':
		return true
	}
	return false
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
//
// 校验规则：
//   - ArtifactID 必填且 > 0：版本列表必须落在某个具体制品库视图内查询。
//   - PackageName 必填且非空白：用于精确定位制品包，缺失时无法返回任何版本。
//
// 注意：此处不在 tag 层使用 binding:required，以便控制器统一抛出多语言错误。
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
// EntityID 只接受 "group.{id}" 或 "version.{id}"。
type ListAssetReq struct {
	EntityID string `json:"entityID" binding:"required"` // group.{id} 或 version.{id}
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Name     string `json:"name"`    // 按文件名模糊过滤
	Version  string `json:"version"` // 按版本号过滤（仅 version 类资产有效）
	OS       string `json:"os"`      // 按操作系统过滤（linux / windows）
	Arch     string `json:"arch"`    // 按 CPU 架构过滤（amd64 / arm64 / arm/v7）
	Creator  string `json:"creator"` // 按创建人过滤
	Editor   string `json:"editor"`  // 按更新人过滤
	Order    string `json:"order"`   // 排序字段：id / createTime / updateTime
	Desc     string `json:"desc"`    // 排序方向：asc / desc
	Deleted  *bool  `json:"deleted"` // 是否包含已删除：nil=不过滤，true=仅已删，false=仅未删
}

// CreateGroupReq 创建 group 节点的请求参数。
// Names 为以 "." 分隔的分组路径，如 "a.b.c" 表示三级分组。
// ArtifactID 必填：必须大于 0。
// Format 不必由前端传入：服务端会根据 artifactID 反查对应视图的 format 并回填，
// 若前端传了也会被覆盖，以保证 group 的 format 与 artifact 绑定一致。
type CreateGroupReq struct {
	ArtifactID int64  `json:"artifactID"`
	Names      string `json:"names" binding:"required"`
	Format     string `json:"format"`
}

// Validate 对 CreateGroupReq 进行业务规则校验，返回携带 i18n key 的错误。
// 校验规则：
//  1. ArtifactID 必须 > 0
//  2. 若 Format 非空，必须通过 enum.ArtifactFormat.Sanitize() 校验（调用方回填 view.Format 后满足）
//  3. Names 按 "." 拆分后不能有空段
//  4. container 格式不支持嵌套（Names 中不能包含 "."）
//
// 注：Format 在控制器阶段会被 view.Format 覆盖，因此这里不再要求前端必传；
// 若调用方直接走 service 入口也可以显式设置 Format 后再调用此方法。
func (r *CreateGroupReq) Validate() error {
	// 0. 兼容 "/" 分隔符：统一替换为 "."
	r.Names = strings.Trim(strings.ReplaceAll(r.Names, "/", "."), ".")

	// 1. ArtifactID 校验
	if r.ArtifactID <= 0 {
		return errors.WithI18nKey(
			fmt.Errorf("artifactID must be greater than 0, got %d", r.ArtifactID),
			i18n.ArtifactGroupArtifactidRequired,
		)
	}

	// 3. Names 合法性校验
	parts := strings.Split(r.Names, ".")
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return errors.WithI18nKey(
				fmt.Errorf("group names contain empty segment: %q", r.Names),
				i18n.ArtifactGroupNamesInvalid,
			)
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

// ParseNames 将 Names 按 "." 拆分为分组路径段。
// 调用前应先通过 Validate() 确保合法性。
func (r *CreateGroupReq) ParseNames() []string {
	return strings.Split(r.Names, ".")
}
