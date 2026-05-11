package artifact

import (
	"fmt"
	"strings"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

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
	SpaceID    int64 `json:"spaceId"`
	RepoID     int64 `json:"repoId"`
	ArtifactID int64 `json:"artifactID"`
}

// SpaceRepoFormReq 定义文件上传场景使用的 Form 请求参数。
type SpaceRepoFormReq struct {
	SpaceID    int64 `form:"spaceId"`
	RepoID     int64 `form:"repoId"`
	ArtifactID int64 `form:"artifactID"`
}

type SoftRemoveEntitiesReq struct {
	EntityIds []string `json:"entity_ids,omitempty"`
}

// ListNodeInfoReq 兼容参考项目 `/artifacts/nodeInfo` 的请求体。
type ListNodeInfoReq struct {
	EntityIds []string `json:"entity_ids,omitempty"`
}

type CreateArtifactViewReq struct {
	SpaceRepoReq
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Format      string `json:"format" binding:"required"`
}

// Validate 对 CreateArtifactViewReq 进行业务规则校验，返回携带 i18n key 的错误。
// 校验规则：
//  1. Format 必须通过 enum.ArtifactFormat.Sanitize() 校验
func (r *CreateArtifactViewReq) Validate() error {

	// 1. Format 枚举校验（使用 constant 包的 Sanitize 模式）
	format := enum.ArtifactFormat(r.Format)
	if _, ok := format.Sanitize(); !ok {
		return errors.WithI18nKey(
			fmt.Errorf("unsupported format: %s", r.Format),
			i18n.ArtifactViewFormatInvalid,
		)
	}

	return nil
}

// ArtifactVersionListReq 接收获取版本列表的请求参数
type ArtifactVersionListReq struct {
	SpaceRepoReq
	PackageName string `json:"packageName" binding:"required"`
	Format      string `json:"format" binding:"required"`
	GroupName   string `json:"groupName"`
}

type ArtifactTreeListReq struct {
	SpaceRepoReq
	ArtifactTreeFilter
}

// CreateGroupReq 创建 group 节点的请求参数。
// Names 为以 "." 分隔的分组路径，如 "a.b.c" 表示三级分组。
// ArtifactID 必填：必须大于 0。
type CreateGroupReq struct {
	ArtifactID int64  `json:"artifactID"`
	Names      string `json:"names" binding:"required"`
	Format     string `json:"format" binding:"required"`
}

// Validate 对 CreateGroupReq 进行业务规则校验，返回携带 i18n key 的错误。
// 校验规则：
//  1. ArtifactID 必须 > 0
//  2. Format 必须通过 enum.ArtifactFormat.Sanitize() 校验
//  3. Names 按 "." 拆分后不能有空段
//  4. container 格式不支持嵌套（Names 中不能包含 "."）
func (r *CreateGroupReq) Validate() error {
	// 1. ArtifactID 校验
	if r.ArtifactID <= 0 {
		return errors.WithI18nKey(
			fmt.Errorf("artifactID must be greater than 0, got %d", r.ArtifactID),
			i18n.ArtifactGroupArtifactidRequired,
		)
	}

	// 2. Format 枚举校验（使用 constant 包的 Sanitize 模式）
	format := enum.ArtifactFormat(r.Format)
	if _, ok := format.Sanitize(); !ok {
		return errors.WithI18nKey(
			fmt.Errorf("unsupported format: %s", r.Format),
			i18n.ArtifactGroupFormatInvalid,
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
	if format == enum.ArtifactFormatContainer && len(parts) > 1 {
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
