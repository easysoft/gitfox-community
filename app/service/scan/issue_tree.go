package scan

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
)

// GetIssueFileTree 获取 issue 文件树
func (s *ScanTaskService) GetIssueFileTree(ctx context.Context, filter scanModel.IssueTreeFilter) (*scanModel.TreeNode, error) {
	taskFilter, err := s.buildIssueTaskFilter(ctx, filter.TaskID)
	if err != nil {
		return nil, err
	}

	res, err := s.issueStore.ListDistinctFilesByRepo(ctx, filter.RepoID, filter.ScanMethods, taskFilter)
	if err != nil {
		return nil, err
	}

	root := &scanModel.TreeNode{Name: "root"}
	for _, raw := range res {
		parts := append([]string{raw.Branch}, strings.Split(raw.File, "/")...)
		insertTreePath(root, parts, raw.File, raw.Count)
	}
	sortTree(root)
	return root, nil
}

// GetIssueRuleTree 获取 issue 规则树
func (s *ScanTaskService) GetIssueRuleTree(ctx context.Context, repoID, taskID int64) (*scanModel.TreeNode, error) {
	taskFilter, err := s.buildIssueTaskFilter(ctx, taskID)
	if err != nil {
		return nil, err
	}

	res, err := s.issueStore.ListDistinctRulesByRepo(ctx, repoID, taskFilter)
	if err != nil {
		return nil, err
	}

	root := &scanModel.TreeNode{Name: "root"}
	for _, raw := range res {
		parts := []string{raw.Plugin, raw.Lang.String(), string(raw.Type), raw.Name}
		insertTreePath(root, parts, raw.ID, raw.Count)
	}
	sortTree(root)
	return root, nil
}

// buildIssueTaskFilter 根据 taskID 通过已有 store 接口构建过滤条件
func (s *ScanTaskService) buildIssueTaskFilter(ctx context.Context, taskID int64) (*scanModel.IssueTaskFilter, error) {
	if taskID == 0 {
		return nil, nil
	}
	task, err := s.taskStore.GetTaskInfo(ctx, uint(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task info: %w", err)
	}
	return &scanModel.IssueTaskFilter{
		TaskID: taskID,
		PlanID: task.PlanID,
		RepoID: task.RepoID,
		Branch: task.ExecutionRef,
	}, nil
}

// insertTreePath 将路径片段插入树中
func insertTreePath(root *scanModel.TreeNode, parts []string, ref any, count int) {
	current := root
	for i, name := range parts {
		found := false
		for _, child := range current.Children {
			if child.Name == name {
				current = child
				found = true
				break
			}
		}
		if !found {
			node := &scanModel.TreeNode{
				Name:   name,
				Path:   filepath.Join(current.Path, name),
				IsLeaf: i == len(parts)-1,
			}
			if node.IsLeaf {
				node.Ref = ref
				node.Count = count
			}
			current.Children = append(current.Children, node)
			current = node
		}
	}
}

// sortTree 递归排序树节点：分支按名称升序，叶子按 count 降序
func sortTree(node *scanModel.TreeNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		ci, cj := node.Children[i], node.Children[j]
		if ci.IsLeaf != cj.IsLeaf {
			return !ci.IsLeaf
		}
		if ci.IsLeaf {
			return ci.Count > cj.Count
		}
		return ci.Name < cj.Name
	})
	for _, child := range node.Children {
		sortTree(child)
	}
}
