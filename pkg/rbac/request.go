// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package rbac

import (
	"fmt"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
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
