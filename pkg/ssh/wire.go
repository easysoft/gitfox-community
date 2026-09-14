// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/git"
	publickeyService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProviderSShServer,
)

func ProviderSShServer(
	cfg *config.Config,
	GitCtrl *git.GitController,
	verifier publickeyService.Interface,
) *Server {
	return &Server{
		Host:              cfg.SSH.Host,
		Port:              cfg.SSH.Port,
		DefaultUser:       cfg.SSH.DefaultUser,
		ServerHostKeysDir: cfg.SSH.HostKeysDir,
		GitCtrl:           GitCtrl,
		Verifier:          verifier,
	}
}
