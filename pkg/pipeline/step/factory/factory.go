// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package factory

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/contracts"
)

var _factory = map[string]contracts.PipelineStepBuildConverter{}

func RegisterConverter(name string, converter contracts.PipelineStepBuildConverter) {
	_factory[name] = converter
}

func GetConverter(name string) (contracts.PipelineStepBuildConverter, error) {
	converter, ok := _factory[name]
	if !ok {
		return nil, fmt.Errorf("project converter not found for %s", name)
	}
	return converter, nil
}
