// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package rbac

import (
	"fmt"

	enum "github.com/easysoft/gitfox-community/app/constant"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

type Request struct {
	Sub    RequestIdentity
	Role   string
	Scope  enum.CasbinScope
	Effect Effect
}

func CompareRequests(want, exist []*Request) (add, remove []*Request) {
	wantMap := make(map[string]*Request)
	existMap := make(map[string]*Request)

	// 构建两个 map,以便比较
	for _, req := range want {
		key := fmt.Sprintf("%v-%s-%s-%s", req.Sub.Key(), req.Role, req.Scope, req.Effect)
		wantMap[key] = req
	}
	for _, req := range exist {
		if req.Effect == EffectDeny {
			continue
		}
		key := fmt.Sprintf("%v-%s-%s-%s", req.Sub.Key(), req.Role, req.Scope, req.Effect)
		existMap[key] = req
	}

	// 比较两个 map,找出差异
	for key, req := range wantMap {
		if _, ok := existMap[key]; !ok {
			add = append(add, req)
			logger.Debugf("set item to add: %s", key)
		}
	}
	for key, req := range existMap {
		if _, ok := wantMap[key]; !ok {
			remove = append(remove, req)
			logger.Debugf("set item to remove: %s", key)
		}
	}

	return
}
