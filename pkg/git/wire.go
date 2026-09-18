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

package git

import (
	"github.com/easysoft/gitfox-community/pkg/cache"
	"github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/git/hook"
	"github.com/easysoft/gitfox-community/pkg/git/storage"
	"github.com/easysoft/gitfox-community/pkg/git/types"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideGitAdapter,
	ProvideGitService,
)

func ProvideGitAdapter(
	config types.Config,
	lastCommitCache cache.Cache[api.CommitEntryKey, *api.Commit],
	githookFactory hook.ClientFactory,
) (*api.Git, error) {
	return api.New(
		config,
		lastCommitCache,
		githookFactory,
	)
}

func ProvideGitService(
	config types.Config,
	adapter *api.Git,
	hookClientFactory hook.ClientFactory,
	storage storage.Store,
) (Interface, error) {
	return New(
		config,
		adapter,
		hookClientFactory,
		storage,
	)
}
