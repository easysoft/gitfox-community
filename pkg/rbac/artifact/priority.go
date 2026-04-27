// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"
)

type Priority int

func (p Priority) String() string          { return fmt.Sprintf("%d", p) }
func (p Priority) Sub(s Priority) Priority { return p - s }

const (
	PrioritySystem   Priority = 999
	PrioritySpace    Priority = 800
	PriorityOwner    Priority = 20
	PriorityPipeline Priority = 10
	PriorityAdmin    Priority = 1

	PriorityNodeInitial Priority = 500
)

type PriorityDecrease int

const (
	PriorityDecreaseDenyRecursiveWildcard Priority = iota
	PriorityDecreaseDenyWildcard
	PriorityDecreaseAllowRecursiveWildcard
	PriorityDecreaseAllowWildcard
)

const PriorityNodeStep Priority = 5

type PriorityNode struct {
	node     string
	priority Priority
}

func NewPriorityNode(node string) (*PriorityNode, error) {
	if strings.HasPrefix(node, "/") || strings.HasSuffix(node, "/.") {
		return nil, fmt.Errorf("invalid path: %s", node)
	}

	frameCount := strings.Count(node, "/") + 1
	n := &PriorityNode{
		node:     node,
		priority: PriorityNodeInitial - Priority(frameCount)*PriorityNodeStep,
	}

	return n, nil
}

func (n PriorityNode) Priority() Priority {
	return n.priority
}

func (n PriorityNode) Identity() rbac.RequestIdentity { return rbac.NewObject(n.node) }

func (n PriorityNode) Wildcard() rbac.RequestIdentity {
	return rbac.NewObject(n.node + "/" + rbac.MatchWildcard)
}

func (n PriorityNode) RecursiveWildcard() rbac.RequestIdentity {
	return rbac.NewObject(n.node + "/" + rbac.MatchRecursiveWildcard)
}
