package repo

import (
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Metadata 获取代码Pull Request 元数据
// @Summary 获取代码Pull Request 元数据
// @Description 根据传入的仓库 ID 和 Pull Request 编号获取 Pull Request 元数据
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pullreqNumber path uint true "Pull Request 编号"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.DiffStats} "获取成功（返回PR的差异统计信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如pullreqNumber格式非法、ID为空等）"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/pullreq/{pullreqNumber}/metadata [get]
func (ctl *Controller) Metadata(ctx *gin.Context) {
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pullreqID, err := GetReqQueryParamID(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	data, err := ctl.pullrequestSvc.GetMetadata(ctx.Request.Context(), session, repoID, pullreqID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PullrequestQuerySuccess, data)
}

// GetDiffStats 获取代码Pull Request 差异统计信息
// @Summary 获取代码Pull Request 差异统计信息
// @Description 根据传入的仓库 ID 和 路径获取 Pull Request 差异统计信息
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param targetPathToSourcePath path string true "目标路径到源路径"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.DiffStats} "获取成功（返回差异统计信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如TargetPathToSourcePath格式非法、为空等）"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/diff-stats/{targetPathToSourcePath} [get]
func (ctl *Controller) GetDiffStats(ctx *gin.Context) {
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pathToPath, err := GetReqQueryParamIDAndPath(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	pathToPath = strings.TrimPrefix(pathToPath, "/")

	data, err := ctl.pullrequestSvc.GetDiffStats(ctx.Request.Context(), session, repoID, pathToPath)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PullrequestQuerySuccess, data)
}

// MergeCheck 获取代码Pull Request 合并检查信息
// @Summary 获取代码Pull Request 合并检查信息
// @Description 根据传入的仓库 ID 和 路径获取 Pull Request 合并检查信息
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param targetPathToSourcePath path string true "目标路径到源路径"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.MergeCheck} "获取成功（返回合并检查信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如TargetPathToSourcePath格式非法、为空等）"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/merge-check/{targetPathToSourcePath} [post]
func (ctl *Controller) MergeCheck(ctx *gin.Context) {
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pathToPath, err := GetReqQueryParamIDAndPath(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	pathToPath = strings.TrimPrefix(pathToPath, "/")

	data, err := ctl.pullrequestSvc.MergeCheck(ctx.Request.Context(), session, repoID, pathToPath)
	if err != nil {
		if errors.GetI18nKey(err) == i18n.GitMergeCheckNoCommits {
			resp.OK(ctx, i18n.PullrequestMergeCheckSuccess, &repoModel.MergeCheck{
				Mergeable: false,
				Message:   i18n.TfGin(ctx, string(i18n.GitMergeCheckNoCommits), nil),
			})
			return
		}
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PullrequestMergeCheckSuccess, data)
}

// DiffSimp 获取代码Pull Request 差异信息
// @Summary 获取代码Pull Request 差异信息
// @Description 根据传入的仓库 ID 和 参数 获取 Pull Request 差异信息
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pullrequestID path uint true "Pull Request 编号"
// @Param Accept header string false "Accept 头，可选值：text/plain, application/json"
// @Description 如果设置为 text/plain，返回 Git 原生 Diff 文本格式；否则返回 JSON 格式
// @Param includePatch query boolean false "是否在返回结果中包含具体的代码补丁 ( patch 字段)。默认为 false 。"
// @Param path query []string false "指定只对比哪些文件（支持多次传递）。"
// @Param range query []string false "指定文件的行号范围 (格式: start:end )，与 path 一一对应。"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.DiffStats} "获取成功（返回差异统计信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如TargetPathToSourcePath格式非法、为空等）"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/pullreq/{pullrequestID}/diff [get]
func (ctl *Controller) DiffSimp(ctx *gin.Context) {
	var param []repoModel.FileDiff
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pullreqID, err := GetReqQueryParamID(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	param, ok = GetFileDiffFromQuery(ctx)
	if !ok {
		resp.BadRequest(ctx, errors.WithI18nKey(nil, i18n.PullrequestPathParamError))
		return
	}

	setSHAs := func(sourceSHA, mergeBaseSHA string) {
		ctx.Writer.Header().Set("X-Source-Sha", sourceSHA)
		ctx.Writer.Header().Set("X-Merge-Base-Sha", mergeBaseSHA)
	}

	if strings.HasPrefix(ctx.Request.Header.Get("Accept"), "text/plain") {
		err := ctl.pullrequestSvc.RawDiff(ctx.Request.Context(), ctx.Writer, session, repoID, pullreqID, setSHAs, param...)
		if err != nil {
			resp.InternalError(ctx, err)
			return
		}
		return
	}

	_, includePatch := req.QueryParam(ctx, "includePatch")
	stream, err := ctl.pullrequestSvc.Diff(ctx.Request.Context(), session, repoID, pullreqID, setSHAs, includePatch, param...)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.JSONArrayDynamic(ctx, ctx.Writer, stream)
}

// Diff 获取代码Pull Request 差异信息
// @Summary 获取代码Pull Request 差异信息
// @Description 根据传入的仓库 ID 和 参数 获取 Pull Request 差异信息
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pullrequestID path uint true "Pull Request 编号"
// @Param Accept header string false "Accept 头，可选值：text/plain, application/json"
// @Description 如果设置为 text/plain，返回 Git 原生 Diff 文本格式；否则返回 JSON 格式
// @Param request body []repoModel.FileDiff true "查询仓库请求参数"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.MergeCheck} "获取成功（返回合并检查信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如TargetPathToSourcePath格式非法、为空等）"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/pullreq/{pullrequestID}/diff [post]
func (ctl *Controller) Diff(ctx *gin.Context) {
	var param []repoModel.FileDiff
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pullreqID, err := GetReqQueryParamID(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindJSON(&param); err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	setSHAs := func(sourceSHA, mergeBaseSHA string) {
		ctx.Writer.Header().Set("X-Source-Sha", sourceSHA)
		ctx.Writer.Header().Set("X-Merge-Base-Sha", mergeBaseSHA)
	}

	if strings.HasPrefix(ctx.Request.Header.Get("Accept"), "text/plain") {
		err := ctl.pullrequestSvc.RawDiff(ctx.Request.Context(), ctx.Writer, session, repoID, pullreqID, setSHAs, param...)
		if err != nil {
			resp.InternalError(ctx, err)
			return
		}
		return
	}

	_, includePatch := req.QueryParam(ctx, "includePatch")
	stream, err := ctl.pullrequestSvc.Diff(ctx.Request.Context(), session, repoID, pullreqID, setSHAs, includePatch, param...)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.JSONArrayDynamic(ctx, ctx.Writer, stream)
}

// Commits 获取代码Pull Request 提交信息
// @Summary 获取代码Pull Request 提交信息
// @Description 根据传入的仓库 ID 和 参数 获取 Pull Request 提交信息
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pullrequestID path uint true "Pull Request 编号"
// @Param page query int false "分页页码，默认1"
// @Param limit query int false "分页每页数量，默认20"
// @Success 200 {object} resp.ListResponse{data=[]commitModel.Commit,listArgs=resp.ListArgs} "查询成功"
// @Error 400 {object} resp.ObjectResponse "参数错误（如TargetPathToSourcePath格式非法、为空等）"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/pullreq/{pullrequestID}/commits [get]
func (ctl *Controller) Commits(ctx *gin.Context) {
	var err error
	var commits []commitModel.Commit

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pullreqID, err := GetReqQueryParamID(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	var filter req.GitListFilter
	if err = ctx.ShouldBindQuery(&filter); err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	req.SetDefaults(&filter)

	commits, totalCommits, err := ctl.pullrequestSvc.Commits(ctx.Request.Context(), session, repoID, pullreqID, uint(filter.Page), uint(filter.Limit))
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, commits,
		resp.ListArgs{
			Total:    int64(totalCommits),
			Page:     int64(filter.Page),
			PageSize: int64(filter.Limit),
		}, i18n.PullrequestQuerySuccess, nil)
}

// Merge 合并代码Pull Request
// @Summary 合并代码Pull Request
// @Description 根据传入的仓库 ID 和 参数 合并 Pull Request
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pullrequestID path uint true "Pull Request 编号"
// @Param request body repoModel.Merge true "合并 Pull Request 请求参数"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.MergeResponse} "合并成功（返回合并信息）"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.MergeViolations} "合并失败（返回合并冲突和规则违规信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如Merge方法、SourceSHA、Title、Message格式非法、为空等）"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/pullreq/{pullrequestID}/merge [post]
func (ctl *Controller) Merge(ctx *gin.Context) {
	var params repoModel.Merge
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pullreqID, err := GetReqQueryParamID(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err = params.Check(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	mergeResp, mergeViolations, err := ctl.pullrequestSvc.Merge(ctx.Request.Context(), session, repoID, pullreqID, params)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	if mergeViolations != nil {
		resp.OK(ctx, i18n.PullrequestMergeViolations, mergeViolations)
		return
	}

	resp.OK(ctx, i18n.PullrequestMergeSuccess, mergeResp)
}

// EventTrigger Pull Request 事件触发
// @Summary Pull Request 创建，关闭，重新打开时调用
// @Description 根据传入的仓库 ID 和 参数 触发 Pull Request 事件
// @Tags pureq
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param pullrequestID path uint true "Pull Request 编号"
// @Param request body repoModel.EventTrigger true "事件触发参数"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.MergeResponse} "触发成功"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.MergeViolations} "触发失败"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id}/pullreq/{pullrequestID}/event-trigger [post]
func (ctl *Controller) EventTrigger(ctx *gin.Context) {
	var params repoModel.EventTrigger
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repoID, pullreqID, err := GetReqQueryParamID(ctx)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 根据 params.Type 触发不同的事件
	err = ctl.pullrequestSvc.TriggerEvent(ctx.Request.Context(), session, repoID, pullreqID, params)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PullrequestEventTriggerSuccess, nil)
}

// GetReqQueryParamID 从路径参数中获取仓库 ID 和 Pull Request 编号
func GetReqQueryParamID(ctx *gin.Context) (uint, uint, error) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		return 0, 0, err
	}
	pullreqID, err := req.PathParamAsPositiveUint(ctx, req.PullrequestID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		return 0, 0, err
	}

	return repoID, pullreqID, nil
}

// GetReqQueryParamIDAndPath 从路径参数中获取仓库 ID 和路径参数
func GetReqQueryParamIDAndPath(ctx *gin.Context) (uint, string, error) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		return 0, "", err
	}

	pathToPath, err := req.PathParamOrError(ctx, req.TargetPathToSourcePath)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PullrequestPathParamError)
		return 0, "", err
	}

	return repoID, pathToPath, nil
}

// GetFileDiffFromQuery 从查询参数中获取文件差异信息
func GetFileDiffFromQuery(ctx *gin.Context) (files []repoModel.FileDiff, ok bool) {
	paths, ok := req.QueryParamList(ctx, "path")
	ranges, ok := req.QueryParamList(ctx, "range")
	if len(paths) != len(ranges) {
		return
	}

	for i, filepath := range paths {
		start := 0
		if i < len(ranges) {
			linesRange := ranges[i]
			parts := strings.Split(linesRange, ":")
			if len(parts) > 0 {
				start, _ = strconv.Atoi(parts[0])
			}
		}
		files = append(files, repoModel.FileDiff{
			Path:      filepath,
			StartLine: start,
		})
	}
	return
}
