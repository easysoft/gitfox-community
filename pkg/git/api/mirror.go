// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox-community/pkg/git/command"
	"github.com/easysoft/gitfox-community/pkg/git/util"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// GetRemoteName returns the name of the remote.
func (g *Git) GetRemoteName() string {
	return "origin"
}

func (g *Git) MirrorAddFetch(ctx context.Context, repoPath, remote, sourceRepo string) error {
	if repoPath == "" {
		return ErrRepositoryPathEmpty
	}
	if remote == "" {
		remote = g.GetRemoteName()
	}
	cmdRemote := command.New("remote",
		command.WithFlag("add"),
		command.WithArg(remote),
		command.WithArg(sourceRepo),
		command.WithFlag("--mirror=fetch"),
	)
	if err := cmdRemote.Run(ctx, command.WithDir(repoPath)); err != nil {
		return err
	}
	logger.GetLogger(ctx).Info("mirror remote added success")
	// typo push miror
	// cmdConfigHead := command.New("config",
	// 	command.WithFlag("--add"),
	// 	command.WithArg("remote."+remote+".push", "+refs/heads/*:refs/heads/*"),
	// )
	// if err := cmdConfigHead.Run(ctx, command.WithDir(repoPath)); err != nil {
	// 	return err
	// }
	// cmdConfigTag := command.New("config",
	// 	command.WithFlag("--add"),
	// 	command.WithArg("remote."+remote+".push", "+refs/tags/*:refs/tags/*"),
	// )
	// if err := cmdConfigTag.Run(ctx, command.WithDir(repoPath)); err != nil {
	// 	return err
	// }
	return nil
}

func (g *Git) MirrorUpdateFetch(ctx context.Context, repoPath, remote, sourceRepo string) error {
	if repoPath == "" {
		return ErrRepositoryPathEmpty
	}
	if remote == "" {
		remote = g.GetRemoteName()
	}
	cmdRemote := command.New("remote",
		command.WithFlag("rm"),
		command.WithArg(remote),
	)
	err := cmdRemote.Run(ctx, command.WithDir(repoPath))
	if err != nil && !strings.HasPrefix(err.Error(), "exit status 128 - fatal: No such remote ") {
		return err
	}
	return g.MirrorAddFetch(ctx, repoPath, remote, sourceRepo)
}

func (g *Git) MirrorFetch(ctx context.Context, repoPath, remote string, prune bool) (error, bool) {
	if repoPath == "" {
		return ErrRepositoryPathEmpty, false
	}
	if remote == "" {
		remote = g.GetRemoteName()
	}
	remoteURL, remoteErr := g.GetRemoteURL(ctx, repoPath, remote)
	if remoteErr != nil {
		logger.GetLogger(ctx).Errorw(
			fmt.Sprintf("SyncMirrors [repo: %v Remote: %v]: GetRemoteAddress Error %v", repoPath, remoteURL, remoteErr),
			"error", remoteErr,
		)
		return remoteErr, false
	}
	cmd := command.New("fetch",
		command.WithArg(remote),
		command.WithFlag("--tags"),
	)
	if prune {
		cmd.Add(command.WithFlag("--prune"))
	}
	// TODO PROXY
	stdoutBuilder := strings.Builder{}
	stderrBuilder := strings.Builder{}
	if err := cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(&stdoutBuilder), command.WithStderr(&stderrBuilder)); err != nil {
		stdout := stdoutBuilder.String()
		stderr := stderrBuilder.String()
		stderrMessage := util.SanitizeCredentialURLs(stderr)
		stdoutMessage := util.SanitizeCredentialURLs(stdout)
		// Now check if the error is a resolve reference due to broken reference
		if strings.Contains(stderr, "unable to resolve reference") && strings.Contains(stderr, "reference broken") {
			logger.GetLogger(ctx).Warnf("SyncMirrors [repo: %v Remote: %v]: failed to update mirror repository due to broken references:\nStdout: %s\nStderr: %s\nErr: %v\nAttempting Prune", repoPath, remoteURL, stdoutMessage, stderrMessage, err)
			err = nil
			// Attempt to prune the repository
			// Attempt prune
			pruneErr := g.pruneBrokenReferences(ctx, repoPath, remote, &stdoutBuilder, &stderrBuilder)
			if pruneErr == nil {
				// Successful prune - reattempt mirror
				stderrBuilder.Reset()
				stdoutBuilder.Reset()
				if err = cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(&stdoutBuilder), command.WithStderr(&stderrBuilder)); err != nil {
					stdout := stdoutBuilder.String()
					stderr := stderrBuilder.String()
					stderrMessage = util.SanitizeCredentialURLs(stderr)
					stdoutMessage = util.SanitizeCredentialURLs(stdout)
				}
			}
		}

		// If there is still an error (or there always was an error)
		if err != nil {
			logger.GetLogger(ctx).Errorw(
				fmt.Sprintf("SyncMirrors [repo: %v Remote: %v]: failed to update mirror repository:\nStdout: %s\nStderr: %s\nErr: %v", repoPath, remoteURL, stdoutMessage, stderrMessage, err),
				"error", err,
			)
			return nil, false
		}
	}
	// TODO 更新size等信息
	return nil, true
}

func (g *Git) pruneBrokenReferences(ctx context.Context,
	repoPath string, remoteName string,
	stdoutBuilder, stderrBuilder *strings.Builder,
) error {
	stderrBuilder.Reset()
	stdoutBuilder.Reset()
	cmd := command.New("remote",
		command.WithFlag("prune"),
		command.WithArg(remoteName),
	)
	pruneErr := cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(stdoutBuilder), command.WithStderr(stderrBuilder))
	if pruneErr != nil {
		stdout := stdoutBuilder.String()
		stderr := stderrBuilder.String()
		// sanitize the output, since it may contain the remote address, which may
		// contain a password
		stderrMessage := util.SanitizeCredentialURLs(stderr)
		stdoutMessage := util.SanitizeCredentialURLs(stdout)
		logger.GetLogger(ctx).Errorw(
			fmt.Sprintf("Failed to prune mirror repository %s references:\nStdout: %s\nStderr: %s\nErr: %v", repoPath, stdoutMessage, stderrMessage, pruneErr),
			"error", pruneErr,
		)
	}
	return pruneErr
}
