// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package stepbuild

import (
	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox-community/pkg/pipeline/step/factory"
)

func ConvertBuild(step *harness.Step) (*harness.StepBuild, error) {
	c, err := factory.GetConverter(step.Type)
	if err != nil {
		return nil, err
	}

	return c.Convert(step)
}
