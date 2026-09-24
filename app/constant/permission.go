// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package constant

// PublicResourceType defines the type of the public resource.
type PublicResourceType string

const (
	PublicResourceTypeRepo  PublicResourceType = "repository"
	PublicResourceTypeSpace PublicResourceType = "space"
)

// PermissionCheck represents a permission check.
type PermissionCheck struct {
	Scope      Scope
	Permission Permission
}

// Resource represents the resource of a permission check.
// Note: Keep the name empty in case access is requested for all resources of that type.
type Resource struct {
	Type       ResourceType
	Identifier string
}

// Scope represents the scope of a permission check
// Notes:
//   - In case the permission check is for resource REPO, keep repo empty (repo is resource, not scope)
//   - In case the permission check is for resource SPACE, SpacePath is an ancestor of the space (space is
//     resource, not scope)
//   - Repo isn't use as of now (will be useful once we add access control for repo child resources, e.g. branches).
type Scope struct {
	ResourceType ResourceType
	ResourceID   uint
}

// ResourceType represents the different types of resources that can be guarded with permissions.
type ResourceType string

const (
	ResourceTypeSpace          ResourceType = "SPACE"
	ResourceTypeRepo           ResourceType = "REPOSITORY"
	ResourceTypeUser           ResourceType = "USER"
	ResourceTypeServiceAccount ResourceType = "SERVICEACCOUNT"
	ResourceTypeService        ResourceType = "SERVICE"
	ResourceTypePipeline       ResourceType = "PIPELINE"
	ResourceTypeSecret         ResourceType = "SECRET"
	ResourceTypeConnector      ResourceType = "CONNECTOR"
	ResourceTypeTemplate       ResourceType = "TEMPLATE"
	ResourceTypeArtifact       ResourceType = "ARTIFACT"
	ResourceTypeGitspace       ResourceType = "GITSPACE"
	ResourceTypeInfraProvider  ResourceType = "INFRAPROVIDER"
	ResourceTypeRegistry       ResourceType = "REGISTRY"
)

// Permission represents the different types of permissions a principal can have.
type Permission string

const (
	/*
	   ----- SPACE -----
	*/
	PermissionSpaceView   Permission = "space_view"
	PermissionSpaceEdit   Permission = "space_edit"
	PermissionSpaceDelete Permission = "space_delete"
)

const (
	/*
		----- REPOSITORY -----
	*/
	PermissionRepoView              Permission = "repo_view"
	PermissionRepoCreate            Permission = "repo_create"
	PermissionRepoEdit              Permission = "repo_edit"
	PermissionRepoDelete            Permission = "repo_delete"
	PermissionRepoPush              Permission = "repo_push"
	PermissionRepoReview            Permission = "repo_review"
	PermissionRepoReportCommitCheck Permission = "repo_reportCommitCheck"
)

const (
	/*
		----- USER -----
	*/
	PermissionUserView      Permission = "user_view"
	PermissionUserEdit      Permission = "user_edit"
	PermissionUserDelete    Permission = "user_delete"
	PermissionUserEditAdmin Permission = "user_editAdmin"
)

const (
	/*
		----- SERVICE ACCOUNT -----
	*/
	PermissionServiceAccountView   Permission = "serviceaccount_view"
	PermissionServiceAccountEdit   Permission = "serviceaccount_edit"
	PermissionServiceAccountDelete Permission = "serviceaccount_delete"
)

const (
	/*
		----- SERVICE -----
	*/
	PermissionServiceView      Permission = "service_view"
	PermissionServiceEdit      Permission = "service_edit"
	PermissionServiceDelete    Permission = "service_delete"
	PermissionServiceEditAdmin Permission = "service_editAdmin"
)

const (
	/*
		----- PIPELINE -----
	*/
	PermissionPipelineView    Permission = "pipeline_view"
	PermissionPipelineEdit    Permission = "pipeline_edit"
	PermissionPipelineDelete  Permission = "pipeline_delete"
	PermissionPipelineExecute Permission = "pipeline_execute"
)

const (
	/*
		----- SECRET -----
	*/
	PermissionSecretView   Permission = "secret_view"
	PermissionSecretEdit   Permission = "secret_edit"
	PermissionSecretDelete Permission = "secret_delete"
	PermissionSecretAccess Permission = "secret_access"
)

const (
	/*
		----- CONNECTOR -----
	*/
	PermissionAIView   Permission = "ai_view"
	PermissionAIEdit   Permission = "ai_edit"
	PermissionAIDelete Permission = "ai_delete"
	PermissionAIAccess Permission = "ai_access"
)

const (
	/*
		----- CONNECTOR -----
	*/
	PermissionConnectorView   Permission = "connector_view"
	PermissionConnectorEdit   Permission = "connector_edit"
	PermissionConnectorDelete Permission = "connector_delete"
	PermissionConnectorAccess Permission = "connector_access"
)

const (
	/*
		----- TEMPLATE -----
	*/
	PermissionTemplateView   Permission = "template_view"
	PermissionTemplateEdit   Permission = "template_edit"
	PermissionTemplateDelete Permission = "template_delete"
	PermissionTemplateAccess Permission = "template_access"
)

const (
	/*
	   ----- ARTIFACTS -----
	*/
	PermissionArtifactView   Permission = "artifact_view"
	PermissionArtifactEdit   Permission = "artifact_edit"
	PermissionArtifactDelete Permission = "artifact_delete"
	PermissionArtifactPush   Permission = "artifact_push"
	PermissionArtifactPull   Permission = "artifact_pull"

	/*
		----- GITSPACE -----
	*/
	PermissionGitspaceView   Permission = "gitspace_view"
	PermissionGitspaceEdit   Permission = "gitspace_edit"
	PermissionGitspaceDelete Permission = "gitspace_delete"
	PermissionGitspaceAccess Permission = "gitspace_access"
)

const (
	/*
		----- INFRAPROVIDER -----
	*/
	PermissionInfraProviderView   Permission = "infraprovider_view"
	PermissionInfraProviderEdit   Permission = "infraprovider_edit"
	PermissionInfraProviderDelete Permission = "infraprovider_delete"
	PermissionInfraProviderAccess Permission = "infraprovider_access"
)

const (
	/*
		----- ARTIFACTS -----
	*/
	PermissionArtifactsDownload Permission = "artifacts_download"
	PermissionArtifactsUpload   Permission = "artifacts_upload"
	PermissionArtifactsDelete   Permission = "artifacts_delete"
)

const (
	/*
		----- REGISTRY -----
	*/
	PermissionRegistryView   Permission = "registry_view"
	PermissionRegistryEdit   Permission = "registry_edit"
	PermissionRegistryDelete Permission = "registry_delete"
)
