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
	"context"
)

// Interface defines the SVN operations used by the application layer.
type Interface interface {
	// Info runs `svn info` to query information about a repository URL or local path.
	Info(ctx context.Context, params *InfoParams) (string, error)

	// InfoRevision is a convenience method that returns the HEAD revision.
	InfoRevision(ctx context.Context, url, username, password string) (string, error)

	// InfoXML returns XML-formatted info for the given URL at the specified revision.
	InfoXML(ctx context.Context, url, revision, username, password string) (string, error)

	// Checkout creates a working copy from the source repository at the target path.
	Checkout(ctx context.Context, params *CheckoutParams) error

	// CheckoutFromRepo creates a working copy from a local SVN repository.
	CheckoutFromRepo(ctx context.Context, repoPath, targetPath string) error

	// Update brings the working copy up to the latest revision.
	Update(ctx context.Context, params *UpdateParams) error

	// CreateRepository creates a new local SVN repository at the given path.
	CreateRepository(ctx context.Context, path string) error

	// RemoveRepository removes the local SVN repository directory.
	RemoveRepository(path string) error

	// RemoveWorkingCopy removes the working copy directory.
	RemoveWorkingCopy(path string) error

	// ResolveRepositoryPaths assembles repository and working copy paths from an encoded directory name.
	ResolveRepositoryPaths(encodedDir string) RepositoryPaths

	// SyncInit initializes svnsync for the target repository.
	SyncInit(ctx context.Context, params *SyncInitParams) error

	// Sync executes svnsync to synchronize revisions.
	Sync(ctx context.Context, params *SyncParams) error

	// SyncCopyRevProps copies revision properties from source to target.
	SyncCopyRevProps(ctx context.Context, params *SyncParams) error
}

// Ensure *Service implements Interface.
var _ Interface = (*Service)(nil)
