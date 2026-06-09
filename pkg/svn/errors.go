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

package svn

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

var (
	ErrSvnBinaryNotFound      = errors.New("svn binary not found")
	ErrSvnVersionNotSupported = errors.New("svn version not supported")
	ErrFailedToGetVersion     = errors.New("failed to get svn version")
	ErrCreateRepository       = errors.New("failed to create svn repository")
	ErrSyncInit               = errors.New("failed to initialize svnsync")
	ErrSync                   = errors.New("failed to sync svnsync")
	ErrSyncCopyRevProps       = errors.New("failed to copy svnsync revprops")
	ErrCheckout               = errors.New("failed to checkout svn working copy")
	ErrUpdate                 = errors.New("failed to update svn working copy")
	ErrRemoveRepository       = errors.New("failed to remove svn repository")
	ErrRemoveWorkingCopy      = errors.New("failed to remove svn working copy")
	ErrNoParamsProvided       = errors.New("params not provided")
)
