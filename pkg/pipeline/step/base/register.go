// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import "github.com/easysoft/gitfox-community/pkg/pipeline/step/factory"

func Register() {
	factory.RegisterConverter("build_go", &buildGoStep{})
	factory.RegisterConverter("build_yarn", &buildYarnStep{})
	factory.RegisterConverter("build_maven", &buildMavenStep{})
	factory.RegisterConverter("build_gradle", &buildGradleStep{})
	factory.RegisterConverter("build_npm", &buildNpmStep{})
}
