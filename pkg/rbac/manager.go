// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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
