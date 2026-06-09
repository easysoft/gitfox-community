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

package command

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

var (
	// ErrInvalidArg represent family of errors to report about bad argument used to make a call.
	ErrInvalidArg = errors.New("invalid argument")
)

// NewError creates an error that wraps the original error and includes stderr output.
// The original error (often *exec.ExitError) is preserved in the error chain and can be
// inspected via errors.Is / errors.As.
func NewError(err error, stderr []byte) error {
	if len(stderr) > 0 {
		return errors.Newf("svn: %s: %s", err.Error(), string(stderr))
	}
	return errors.Newf("svn: %s", err.Error())
}
