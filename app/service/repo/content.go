package repo

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/utils"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox/types/enum"
	"github.com/rs/zerolog/log"
)

func (s *Service) GetContent(ctx context.Context, session *auth.Session, repo *repoModel.Repo, gitRef string, repoPath string, includeLatestCommit bool, flattenDirectories bool) (*repoModel.GetContentOutput, error) {

	// create read params once
	readParams := git.ReadParams{
		RepoUID: repo.GitUID,
	}
	treeNodeOutput, err := s.gitSvc.GetTreeNode(ctx, &git.GetTreeNodeParams{
		ReadParams:          readParams,
		GitREF:              gitRef,
		Path:                repoPath,
		IncludeLatestCommit: includeLatestCommit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tree node: %s", repoPath)
	}

	info, err := mapToContentInfo(treeNodeOutput.Node, treeNodeOutput.Commit, includeLatestCommit)
	if err != nil {
		return nil, err
	}

	var content repoModel.Content
	switch info.Type {
	case repoModel.ContentTypeDir:
		content, err = s.getDirContent(ctx, readParams, gitRef, repoPath, flattenDirectories)
	case repoModel.ContentTypeFile:
		content, err = s.getFileContent(ctx, readParams, info.SHA)
	case repoModel.ContentTypeSymlink:
		content, err = s.getSymlinkContent(ctx, readParams, info.SHA)
	case repoModel.ContentTypeSubmodule:
		content, err = s.getSubmoduleContent(ctx, readParams, gitRef, repoPath, info.SHA)
	default:
		err = fmt.Errorf("unknown tree node type '%s'", treeNodeOutput.Node.Type)
	}

	if err != nil {
		return nil, err
	}

	return &repoModel.GetContentOutput{
		ContentInfo: info,
		Content:     content,
	}, nil
}

func (s *Service) getSubmoduleContent(ctx context.Context,
	readParams git.ReadParams,
	gitRef string,
	repoPath string,
	commitSHA string,
) (*repoModel.SubmoduleContent, error) {
	output, err := s.gitSvc.GetSubmodule(ctx, &git.GetSubmoduleParams{
		ReadParams: readParams,
		GitREF:     gitRef,
		Path:       repoPath,
	})
	if err != nil {
		// TODO: handle not found error
		return nil, fmt.Errorf("failed to get submodule: %w", err)
	}

	return &repoModel.SubmoduleContent{
		URL:       output.Submodule.URL,
		CommitSHA: commitSHA,
	}, nil
}

func (s *Service) getFileContent(ctx context.Context,
	readParams git.ReadParams,
	blobSHA string,
) (*repoModel.FileContent, error) {
	output, err := s.gitSvc.GetBlob(ctx, &git.GetBlobParams{
		ReadParams: readParams,
		SHA:        blobSHA,
		SizeLimit:  repoModel.MaxGetContentFileSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	defer func() {
		if err := output.Content.Close(); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msgf("failed to close blob content reader.")
		}
	}()

	content, err := io.ReadAll(output.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob content: %w", err)
	}

	return &repoModel.FileContent{
		Size:     output.Size,
		DataSize: output.ContentSize,
		Encoding: enum.ContentEncodingTypeBase64,
		Data:     base64.StdEncoding.EncodeToString(content),
	}, nil
}

func (s *Service) getSymlinkContent(ctx context.Context,
	readParams git.ReadParams,
	blobSHA string,
) (*repoModel.SymlinkContent, error) {
	output, err := s.gitSvc.GetBlob(ctx, &git.GetBlobParams{
		ReadParams: readParams,
		SHA:        blobSHA,
		SizeLimit:  repoModel.MaxGetContentFileSize, // TODO: do we need to guard against too big symlinks?
	})
	if err != nil {
		// TODO: handle not found error
		return nil, fmt.Errorf("failed to get symlink: %w", err)
	}

	defer func() {
		if err := output.Content.Close(); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msgf("failed to close blob content reader.")
		}
	}()

	content, err := io.ReadAll(output.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob content: %w", err)
	}

	return &repoModel.SymlinkContent{
		Size:   output.Size,
		Target: string(content),
	}, nil
}

func (s *Service) getDirContent(ctx context.Context,
	readParams git.ReadParams,
	gitRef string,
	repoPath string,
	flattenDirectories bool,
) (*repoModel.DirContent, error) {
	output, err := s.gitSvc.ListTreeNodes(ctx, &git.ListTreeNodeParams{
		ReadParams:         readParams,
		GitREF:             gitRef,
		Path:               repoPath,
		FlattenDirectories: flattenDirectories,
	})
	if err != nil {
		// TODO: handle not found error
		return nil, fmt.Errorf("failed to get content of dir: %w", err)
	}

	entries := make([]repoModel.ContentInfo, len(output.Nodes))
	for i, node := range output.Nodes {
		entries[i], err = mapToContentInfo(node, nil, false)
		if err != nil {
			return nil, err
		}
	}

	return &repoModel.DirContent{
		Entries: entries,
	}, nil
}

// mapToContentInfo maps git.TreeNode to repoModel.ContentInfo
// It maps the node mode to the content type and sets the SHA, Name, and Path fields.
// If includeLatestCommit is true and commit is not nil, it maps the commit to the LatestCommit field.
// 映射git.TreeNode到repoModel.ContentInfo
// 它将节点模式映射到内容类型并设置SHA、名称和路径字段。
// 如果includeLatestCommit为true且commit不为nil，它将提交映射到LatestCommit字段。
func mapToContentInfo(node git.TreeNode, commit *git.Commit, includeLatestCommit bool) (repoModel.ContentInfo, error) {
	typ, err := mapNodeModeToContentType(node.Mode)
	if err != nil {
		return repoModel.ContentInfo{}, err
	}

	res := repoModel.ContentInfo{
		Type: typ,
		SHA:  node.SHA,
		Name: node.Name,
		Path: node.Path,
	}

	// parse commit only if available
	if commit != nil && includeLatestCommit {
		res.LatestCommit, err = utils.MapCommit(commit)
		if err != nil {
			return repoModel.ContentInfo{}, err
		}
	}

	return res, nil
}

// mapNodeModeToContentType maps git.TreeNodeMode to repoModel.ContentType
// It returns an error if the mode is not supported.
// 映射git.TreeNodeMode到repoModel.ContentType
// 如果模式不是文件、执行文件、符号链接、提交或目录，返回错误。
func mapNodeModeToContentType(m git.TreeNodeMode) (repoModel.ContentType, error) {
	switch m {
	case git.TreeNodeModeFile, git.TreeNodeModeExec:
		return repoModel.ContentTypeFile, nil
	case git.TreeNodeModeSymlink:
		return repoModel.ContentTypeSymlink, nil
	case git.TreeNodeModeCommit:
		return repoModel.ContentTypeSubmodule, nil
	case git.TreeNodeModeTree:
		return repoModel.ContentTypeDir, nil
	default:
		return repoModel.ContentTypeFile, errors.InternalError.Newf("unsupported tree node mode '%s'", m)
	}
}
