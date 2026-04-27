// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package rbac

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type Manager struct {
	Enf casbin.IEnforcer
}

func NewManager(db *gorm.DB, table string, m model.Model) (*Manager, error) {
	a, err := gormadapter.NewAdapterByDBUseTableName(db, "rbac", table)
	if err != nil {
		return nil, err
	}

	enf, err := casbin.NewSyncedEnforcer(m, a)
	if err != nil {
		return nil, err
	}

	if err = enf.LoadPolicy(); err != nil {
		return nil, err
	}

	return &Manager{Enf: enf}, nil
}

func (m *Manager) AddPolicies(data ...Policy) (bool, error) {
	rules := make([][]string, len(data))
	for i, p := range data {
		rules[i] = p.Values()
	}
	return m.Enf.AddPolicies(rules)
}

func (m *Manager) RemovePolicies(data ...Policy) (bool, error) {
	rules := make([][]string, len(data))
	for i, p := range data {
		rules[i] = p.Values()
	}
	return m.Enf.RemovePolicies(rules)
}

func (m *Manager) RemoveFilteredPolicy(index PolicyItemIndex, item ...string) (bool, error) {
	return m.Enf.RemoveFilteredPolicy(int(index), item...)
}
