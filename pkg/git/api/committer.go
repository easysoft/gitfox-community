// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/git/command"
	"github.com/easysoft/gitfox-community/pkg/utils/common"

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
