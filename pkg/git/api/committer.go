// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/command"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/common"

	"github.com/golang-module/carbon/v2"
	"golang.org/x/exp/slices"
)

// ListCommitters lists the committers reachable from ref.
func (g *Git) ListCommitters(
	ctx context.Context,
	repoPath string,
	ref string,
	filter CommitterFilter,
) ([]string, error) {
	if repoPath == "" {
		return nil, ErrRepositoryPathEmpty
	}
	cmd := command.New("log",
		command.WithFlag("--pretty=format:"+fmtAuthorEmail),
		command.WithFlag("--no-merges"),
	)
	if len(filter.Since) == 10 {
		since := common.DayStartUnixRFC3339(carbon.Parse(filter.Since))
		cmd.Add(command.WithFlag("--since", since))
	}
	if len(filter.Until) == 10 {
		until := common.DayEndUnixRFC3339(carbon.Parse(filter.Until))
		cmd.Add(command.WithFlag("--until", until))
	}

	if len(ref) == 0 {
		cmd.Add(command.WithFlag("--branches=*"))
	} else {
		cmd.Add(command.WithFlag("--first-parent", ref))
	}
	output := &bytes.Buffer{}
	err := cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(output))
	if err != nil {
		return nil, fmt.Errorf("failed to run git to get committers: %w", err)
	}
	scanner := bufio.NewScanner(output)
	var emails []string
	for scanner.Scan() {
		email := strings.TrimSpace(scanner.Text())
		if len(email) > 0 && !slices.Contains(emails, email) {
			emails = append(emails, email)
		}
	}

	return emails, nil
}
