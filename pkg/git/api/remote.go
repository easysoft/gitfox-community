// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package api

import (
	"bytes"
	"context"
	"strings"

	giturl "github.com/easysoft/gitfox-community/pkg/git/api/url"
	"github.com/easysoft/gitfox-community/pkg/git/command"
)

// GetRemoteAddress returns remote url of git repository in the repoPath with special remote name
func (g *Git) GetRemoteAddress(ctx context.Context, repoPath, remoteName string) (string, error) {
	// > 2.7 git remote get-url origin
	// < 2.7 git config --get remote.origin.url

	cmd := command.New("remote",
		command.WithFlag("get-url"),
		command.WithArg(remoteName),
	)
	output := &bytes.Buffer{}

	err := cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(output))
	if err != nil {
		return "", err
	}
	result := strings.TrimSpace(output.String())
	if len(result) > 0 {
		result = result[:len(result)-1]
	}
	return result, nil
}

// GetRemoteURL returns the url of a specific remote of the repository.
func (g *Git) GetRemoteURL(ctx context.Context, repoPath, remoteName string) (*giturl.GitURL, error) {
	addr, err := g.GetRemoteAddress(ctx, repoPath, remoteName)
	if err != nil {
		return nil, err
	}
	return giturl.Parse(addr)
}
