// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact_test

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac/artifact"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRole(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})

	m, err := artifact.NewManager(db)
	require.NoError(t, err)

	err = m.RM.EnsureDefaultRoles()
	require.NoError(t, err)

	_, err = m.RM.ListAllRoles()
	require.NoError(t, err)

	space := rbac.NewSpace(1)
	user := rbac.NewUser(1)
	user2 := rbac.NewUser(2)

	ctx := context.Background()

	err = m.PM.DenyAllAccessToNode(space, rbac.NewObject("raw/group:1/group2"))
	require.NoError(t, err)

	err = m.PM.GrantRoleToUserOnNode(user, space, rbac.NewObject("raw/group:1/group2"), enum.CasbinArtifactRoleReader, enum.CasbinScopeRecursive)
	require.NoError(t, err)

	err = m.PM.GrantRoleToUserOnNode(user2, space, rbac.NewObject("raw/group:1/group2"), enum.CasbinArtifactRoleReader, enum.CasbinScopeSub)
	require.NoError(t, err)

	tests := []struct {
		user     rbac.RequestIdentity
		space    rbac.RequestIdentity
		path     string
		action   enum.CasbinAction
		expected bool
	}{
		{user, space, "raw/group:1/group2/a/b.txt", enum.CasbinActionArtifactDelete, false},
		{user2, space, "raw/group:1/group2/a/b.txt", enum.CasbinActionArtifactDelete, false},
		{user, space, "raw/group:1/group2/a/b.txt", enum.CasbinActionArtifactPull, true},
		{user2, space, "raw/group:1/group2/a/b.txt", enum.CasbinActionArtifactPull, false},
		{user, space, "raw/group:1/group2/index.html", enum.CasbinActionArtifactPull, true},
		{user2, space, "raw/group:1/group2/index.html", enum.CasbinActionArtifactPull, true},
		{user, space, "raw/group:1/group2", enum.CasbinActionArtifactView, true},
		{user2, space, "raw/group:1/group2", enum.CasbinActionArtifactView, false},
	}

	for idx, test := range tests {
		success, reason, _ := m.Enforce(ctx, test.user, test.space, rbac.NewObject(test.path), test.action)
		if test.expected {
			require.True(t, success, "idx: %d, request: %+v,reason: %+v", idx, test, reason)
		} else {
			require.False(t, success, "idx: %d, request: %+v,reason: %+v", idx, test, reason)
		}
	}

	//require.Equal(t, 1, 2)
}
