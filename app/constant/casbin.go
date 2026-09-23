// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package constant

type CasbinAction string

func (s CasbinAction) String() string {
	return string(s)
}

const (
	CasbinActionArtifactView     CasbinAction = "view"
	CasbinActionArtifactPull     CasbinAction = "pull"
	CasbinActionArtifactPush     CasbinAction = "push"
	CasbinActionArtifactOverride CasbinAction = "override"

	// CasbinActionArtifactCreate is used for create group / package
	CasbinActionArtifactCreate CasbinAction = "create"
	CasbinActionArtifactUpdate CasbinAction = "update"
	CasbinActionArtifactDelete CasbinAction = "delete"
	CasbinActionPermission     CasbinAction = "permission"
)

type CasbinRole string

func (s CasbinRole) String() string {
	return string(s)
}

var CasbinRoles = []CasbinRole{
	CasbinArtifactRoleReader,
	CasbinArtifactRoleWriter,
	CasbinArtifactRoleAdmin,
}

const (
	// view pull
	CasbinArtifactRoleReader CasbinRole = "artifact_reader"
	// view pull push
	CasbinArtifactRoleWriter CasbinRole = "artifact_writer"
	// view pull push override update delete create permission
	CasbinArtifactRoleAdmin CasbinRole = "artifact_admin"
)

type CasbinScope string

func (s CasbinScope) String() string { return string(s) }

var CasbinScopes = []CasbinScope{
	CasbinScopeSingle,
	CasbinScopeSub,
	CasbinScopeRecursive,
}

const (
	// 当前节点权限
	CasbinScopeSingle CasbinScope = "."
	// 子节点权限
	CasbinScopeSub CasbinScope = "*"
	// 递归权限
	CasbinScopeRecursive CasbinScope = "**"
)
