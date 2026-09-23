// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package api

import (
	"bytes"
	"context"
	"strconv"

	"github.com/easysoft/gitfox-community/pkg/git/command"
)

var excludeFromDiff = []string{
	"package-lock.json",
	"pnpm-lock.yaml",
	// yarn.lock, Cargo.lock, Gemfile.lock, Pipfile.lock, etc.
	"*.lock",
	"go.sum",
}

func (g *Git) excludeFiles(excludeList []string) []string {
	var excludedFiles []string
	// TODO: add more exclude files
	excludeList = append(excludeList, excludeFromDiff...)
	for _, f := range excludeList {
		excludedFiles = append(excludedFiles, ":(exclude,top)"+f)
	}
	return excludedFiles
}

func (g *Git) DiffFiles(ctx context.Context,
	repoPath string,
	baseRef string,
	headRef string,
	unified int,
	excludeList []string,
) (string, error) {
	cmd := command.New("diff",
		command.WithFlag("--ignore-all-space"),
		command.WithFlag("--diff-algorithm=minimal"),
		// command.WithFlag("--diff-filter=ad"),
		command.WithFlag("--unified="+strconv.Itoa(unified)),
	)
	cmd.Add(command.WithArg(baseRef, headRef))
	cmd.Add(command.WithArg(g.excludeFiles(excludeList)...))
	stdout := &bytes.Buffer{}
	err := cmd.Run(ctx,
		command.WithDir(repoPath),
		command.WithStdout(stdout),
	)
	if err != nil {
		return "", processGitErrorf(err, "failed to trigger diff command")
	}

	return stdout.String(), nil
}
