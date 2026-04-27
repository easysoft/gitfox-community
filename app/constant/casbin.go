// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

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
	CasbinArtifactRoleReader CasbinRole = "artifact_reader"
	CasbinArtifactRoleWriter CasbinRole = "artifact_writer"
	CasbinArtifactRoleAdmin  CasbinRole = "artifact_admin"
)

type CasbinScope string

func (s CasbinScope) String() string    { return string(s) }

var CasbinScopes = []CasbinScope{
	CasbinScopeSingle,
	CasbinScopeSub,
	CasbinScopeRecursive,
}

const (
	CasbinScopeSingle    CasbinScope = "."
	CasbinScopeSub       CasbinScope = "*"
	CasbinScopeRecursive CasbinScope = "**"
)
