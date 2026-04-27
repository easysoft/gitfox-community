package artifact

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

// SpaceRepoReq 定义接收含有 spaceId 和 repoId 参数的通用请求。
type SpaceRepoReq struct {
	SpaceID int64 `json:"spaceId" form:"spaceId"`
	RepoID  int64 `json:"repoId" form:"repoId"`
}

type SoftRemoveEntitiesReq struct {
	EntityIds []string `json:"entity_ids,omitempty"`
	NodeIds   []string `json:"node_ids,omitempty"`
}

type CreateArtifactViewReq struct {
	SpaceRepoReq
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}


// ArtifactVersionListReq 接收获取版本列表的请求参数
type ArtifactVersionListReq struct {
	SpaceRepoReq
	PackageName string `json:"packageName" binding:"required"`
	Format      string `json:"format" binding:"required"`
	GroupName   string `json:"groupName"`
}
