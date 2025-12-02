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

package githook

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
)

var (
	// ExecutionTimeout is the timeout used for githook CLI runs.
	ExecutionTimeout = 3 * time.Minute
)

// GenerateEnvironmentVariables generates the required environment variables for a payload
// constructed from the provided parameters.
// The parameter `internal` should be true if the call is coming from the Harness
// and therefore protection from rules shouldn't be verified.
func GenerateEnvironmentVariables(
	ctx context.Context,
	apiBaseURL string,
	repoID int64,
	principalAccount string,
	disabled bool,
	internal bool,
) (map[string]string, error) {
	// 从 ctx 拿 requestID
	// generate githook base url
	baseURL := strings.TrimLeft(apiBaseURL, "/") + "/v2/internal/git-hooks"

	payload := Payload{
		BaseURL:          baseURL,
		RepoID:           repoID,
		PrincipalAccount: principalAccount,
		RequestID:        "requestID",
		Disabled:         disabled,
		Internal:         internal,
	}

	if err := payload.Validate(); err != nil {
		return nil, fmt.Errorf("generated payload is invalid: %w", err)
	}

	return hook.GenerateEnvironmentVariables(payload)
}

// LoadFromEnvironment returns a new githook.CLICore created by loading the payload from the environment variable.
func LoadFromEnvironment() (*hook.CLICore, error) {
	payload, err := hook.LoadPayloadFromEnvironment[Payload]()
	if err != nil {
		return nil, fmt.Errorf("failed to load payload from environment: %w", err)
	}

	// ensure we return disabled error in case it's explicitly disabled (will result in no-op)
	if payload.Disabled {
		return nil, hook.ErrDisabled
	}

	if err := payload.Validate(); err != nil {
		return nil, fmt.Errorf("payload validation failed: %w", err)
	}

	return hook.NewCLICore(
		NewRestClient(payload),
		ExecutionTimeout,
	), nil
}
