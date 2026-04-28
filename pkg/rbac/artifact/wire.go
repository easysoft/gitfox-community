// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideArtifactPermManager,
)

// ProvideArtifactPermManager provides a casbin perm manager
func ProvideArtifactPermManager(
	db *gorm.DB,
) (*Manager, error) {
	m, err := NewManager(db)
	if err != nil {
		return nil, err
	}

	// 初始化角色-动作映射（g2 表）：artifact_reader/writer/admin 对应的 action 集合。
	// 必须在服务启动时调用，否则 Casbin g2 映射缺失，所有基于角色的权限校验均失败。
	if err = m.RM.EnsureDefaultRoles(); err != nil {
		return nil, err
	}

	// 初始化系统全局默认规则：* * ** artifact_reader allow priority=999
	// 确保无任何自定义策略时，已登录用户默认拥有 reader 权限。
	if err = m.PM.ensureSystemDefaultRule(); err != nil {
		return nil, err
	}

	return m, nil
}
