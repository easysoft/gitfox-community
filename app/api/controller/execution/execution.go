package execution

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/canceler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	executionService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	pipelineService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	repoService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/livelog"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

var (
	pingInterval = 30 * time.Second
	tailMaxTime  = 1 * time.Hour
)

// logLineWriter 定义日志行的写入方式
type logLineWriter func(w io.Writer, line *livelog.Line)

// writeLineAsJSON 以 JSON 格式写入日志行（用于 GetLiveLogs）
func writeLineAsJSON(w io.Writer, line *livelog.Line) {
	io.WriteString(w, "data: ")
	json.NewEncoder(w).Encode(line)
	io.WriteString(w, "\n\n")
}

// writeLineAsText 以纯文本格式写入日志行（用于 GetLogStream）
func writeLineAsText(w io.Writer, line *livelog.Line) {
	ts := time.Unix(line.Timestamp, 0).Format("2006-01-02 15:04:05")
	fmt.Fprintf(w, "data: [%s] %s\n\n", ts, line.Message)
}

// parseLogPathParams 解析日志相关的四个路径参数
func parseLogPathParams(ctx *gin.Context) (pipelineID, executionID, stageID, stepID uint, ok bool) {
	var err error
	pipelineID, err = req.PathParamAsPositiveUint(ctx, req.PathParamPipelineID)
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ExecutionPathParamError))
		return 0, 0, 0, 0, false
	}
	executionID, err = req.PathParamAsPositiveUint(ctx, req.PathParamExecutionID)
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ExecutionPathParamError))
		return 0, 0, 0, 0, false
	}
	stageID, err = req.PathParamAsPositiveUint(ctx, req.PathParamStageID)
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ExecutionPathParamError))
		return 0, 0, 0, 0, false
	}
	stepID, err = req.PathParamAsPositiveUint(ctx, req.PathParamStepID)
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ExecutionPathParamError))
		return 0, 0, 0, 0, false
	}
	return pipelineID, executionID, stageID, stepID, true
}

// buildS3LogsParam 构建文件日志存储参数
func buildS3LogsParam(pipelineID, executionID, stageID, stepID uint) *pipelineModel.S3Logs {
	return &pipelineModel.S3Logs{
		PipelineID:  strconv.FormatUint(uint64(pipelineID), 10),
		ExecutionID: strconv.FormatUint(uint64(executionID), 10),
		StageID:     strconv.FormatUint(uint64(stageID), 10),
		StepID:      strconv.FormatUint(uint64(stepID), 10),
	}
}

// setupSSE 设置 SSE 响应头并返回 Flusher，失败时直接写入错误响应并返回 nil
func setupSSE(ctx *gin.Context) (http.Flusher, bool) {
	w := ctx.Writer
	f, ok := w.(http.Flusher)
	if !ok {
		log := logger.GetLoggerFromGin(ctx)
		log.Errorf("http writer does not support flushing")
		resp.InternalError(ctx, errors.WithI18nKey(nil, i18n.ExecutionLogQueryError))
		return nil, false
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	h.Set("Access-Control-Allow-Origin", "*")

	io.WriteString(w, ": ping\n\n")
	f.Flush()
	return f, true
}

// tailLiveStream 消费实时日志流并通过 SSE 写出，直到流结束或超时
func (c *Controller) tailLiveStream(ctx *gin.Context, stepID uint, f http.Flusher, writeLine logLineWriter) {
	log := logger.GetLoggerFromGin(ctx)
	w := ctx.Writer

	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), tailMaxTime)
	defer cancel()

	linec, errc := c.pipelineService.TailLogStream(reqCtx, int64(stepID))

	pingTimer := time.NewTimer(pingInterval)
	defer pingTimer.Stop()

	for {
		if !pingTimer.Stop() {
			select {
			case <-pingTimer.C:
			default:
			}
		}
		pingTimer.Reset(pingInterval)
		select {
		case <-reqCtx.Done():
			io.WriteString(w, "event: error\ndata: eof\n\n")
			f.Flush()
			return
		case err := <-errc:
			log.Errorf("received error in the tail channel: %v", err)
			io.WriteString(w, "event: error\ndata: eof\n\n")
			f.Flush()
			return
		case <-pingTimer.C:
			io.WriteString(w, ": ping\n\n")
			f.Flush()
		case line := <-linec:
			writeLine(w, line)
			f.Flush()
		}
	}
}

// Controller 处理与 pipeline 执行相关的 HTTP 请求
type Controller struct {
	triggerer        triggerer.Triggerer
	canceler         canceler.Canceler
	executionService executionService.Interface
	pipelineService  pipelineService.Interface
	repoService      repoService.Interface
	gitPkg           git.Interface
}

// NewController 创建流水线执行实例
func NewController(
	triggerer triggerer.Triggerer,
	canceler canceler.Canceler,
	executionService executionService.Interface,
	pipelineService pipelineService.Interface,
	repoService repoService.Interface,
	gitPkg git.Interface,
) *Controller {
	return &Controller{
		triggerer:        triggerer,
		canceler:         canceler,
		executionService: executionService,
		pipelineService:  pipelineService,
		repoService:      repoService,
		gitPkg:           gitPkg,
	}
}

// @Summary 创建流水线执行
// @Description 创建新的流水线执行
// @Tags executions
// @Accept application/json
// @Produce application/json
// @Param request body executionModel.ReqForCreate true "创建流水线执行的请求参数"
// @Success 200 {object} resp.ObjectResponse{data=executionModel.Execution} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/executions [post]
func (c *Controller) Create(ctx *gin.Context) {
	var reqBody executionModel.ReqForCreate
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	ref, ok := reqBody.Params["gitRef"]
	if !ok {
		ref = ""
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	// 获取 pipeline 关联的 repo 信息
	pipeline, err := c.pipelineService.GetPipelineByID(ctx, int64(reqBody.PipelineID))
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 获取分支信息以取得 SHA
	var commitSHA string
	if ref != "" && pipeline.RepoID != 0 {
		repo, err := c.repoService.GetByID(ctx, session, uint(pipeline.RepoID))
		if err != nil {
			resp.InternalError(ctx, err)
			return
		}
		branchOut, err := c.gitPkg.GetBranch(ctx, &git.GetBranchParams{
			ReadParams: git.ReadParams{RepoUID: repo.GitUID},
			BranchName: ref,
		})
		if err != nil {
			resp.InternalError(ctx, err)
			return
		}
		commitSHA = branchOut.Branch.SHA.String()
	}

	// Create manual hook for execution.
	hook := &triggerer.Hook{
		Trigger:     session.User.Account, // who/what triggered the build, different from commit author
		TriggeredBy: session.User.ID,
		Parent:      pipeline.SpaceID,
		Ref:         ref,
		Commit:      commitSHA,
		Sender:      session.User.Account,
		Params:      reqBody.Params,
	}

	executionOutput, err := c.triggerer.Trigger(ctx, reqBody.PipelineID, hook)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionTriggerError)
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ExecutionTriggerSuccess, executionOutput)
}

// @Summary 查询流水线执行详情
// @Description 查询流水线执行的详细信息
// @Tags executions
// @Accept application/json
// @Produce application/json
// @Param id path uint true "流水线执行ID"
// @Success 200 {object} resp.ObjectResponse{data=executionModel.ExecutionWithStages} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/executions/{id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	result, err := c.executionService.GetWithStages(ctx, id)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionQueryError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.ExecutionQuerySuccess, result)
}

// @Summary 取消流水线执行
// @Description 取消指定的流水线执行
// @Tags executions
// @Accept application/json
// @Produce application/json
// @Param id path uint true "流水线执行ID"
// @Success 200 {object} resp.ObjectResponse "取消成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/executions/{id}/cancel [post]
func (c *Controller) Cancel(ctx *gin.Context) {

	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	err = c.canceler.Cancel(ctx, id, session.User.Account)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionCancelError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.ExecutionCancelSuccess, nil)
}

// @Summary 流式获取流水线执行日志
// @Description 以 SSE 方式流式返回指定步骤的实时日志流
// @Tags executions
// @Accept application/json
// @Produce text/event-stream
// @Param pipelineID path string true "流水线ID"
// @Param executionID path string true "执行ID"
// @Param stageID path string true "阶段ID"
// @Param stepID path string true "步骤ID"
// @Success 200 {string} string "SSE 日志流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/executions/{executionID}/stream/{stageID}/{stepID} [get]
func (c *Controller) GetLiveLogs(ctx *gin.Context) {
	stepID, err := req.PathParamAsPositiveUint(ctx, req.PathParamStepID)
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ExecutionPathParamError))
		return
	}

	f, ok := setupSSE(ctx)
	if !ok {
		return
	}

	c.tailLiveStream(ctx, stepID, f, writeLineAsJSON)
}

// @Summary 创建实时日志流
// @Description 为指定步骤创建实时日志流
// @Tags executions
// @Accept application/json
// @Produce application/json
// @Param pipelineID path string true "流水线ID"
// @Param executionID path string true "执行ID"
// @Param stageID path string true "阶段ID"
// @Param stepID path string true "步骤ID"
// @Success 200 {object} resp.ObjectResponse "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/executions/{executionID}/stream/{stageID}/{stepID} [post]
func (c *Controller) CreateLiveLogs(ctx *gin.Context) {
	stepID, err := req.PathParamAsPositiveUint(ctx, req.PathParamStepID)
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ExecutionPathParamError))
		return
	}

	sid := int64(stepID)

	err = c.pipelineService.CreateLogStream(ctx, sid)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionLogCreateError)
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 异步写入模拟日志，方便测试 tail 端实时接收
	go func() {
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Second)
			_ = c.pipelineService.WriteLogStream(context.Background(), sid, &livelog.Line{
				Number:    i,
				Message:   fmt.Sprintf("[step:%d] log line %d — pipeline is running...", sid, i),
				Timestamp: time.Now().Unix(),
			})
		}
		// 写完后删除流，tail 端会收到 EOF
		_ = c.pipelineService.DeleteLogStream(context.Background(), sid)
	}()

	resp.OK(ctx, i18n.ExecutionLogCreateSuccess, nil)
}

// @Summary 获取文件日志
// @Description 从文件存储中获取指定步骤的执行日志
// @Tags executions
// @Produce text/plain
// @Param pipelineID path string true "流水线ID"
// @Param executionID path string true "执行ID"
// @Param stageID path string true "阶段ID"
// @Param stepID path string true "步骤ID"
// @Success 200 {string} string "日志内容"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/executions/{executionID}/logs/{stageID}/{stepID} [get]
func (c *Controller) GetFileLogs(ctx *gin.Context) {
	pipelineID, executionID, stageID, stepID, ok := parseLogPathParams(ctx)
	if !ok {
		return
	}

	rc, err := c.pipelineService.GetLogs(ctx, buildS3LogsParam(pipelineID, executionID, stageID, stepID))
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionLogQueryError)
		resp.InternalError(ctx, err)
		return
	}
	defer rc.Close()

	ctx.Header("Content-Type", "text/plain; charset=utf-8")
	io.Copy(ctx.Writer, rc)
}

// @Summary 上传文件日志
// @Description 将请求体内容作为日志写入文件存储
// @Tags executions
// @Accept text/plain
// @Produce application/json
// @Param pipelineID path string true "流水线ID"
// @Param executionID path string true "执行ID"
// @Param stageID path string true "阶段ID"
// @Param stepID path string true "步骤ID"
// @Success 200 {object} resp.ObjectResponse "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/executions/{executionID}/logs/{stageID}/{stepID} [post]
func (c *Controller) CreateFileLogs(ctx *gin.Context) {
	pipelineID, executionID, stageID, stepID, ok := parseLogPathParams(ctx)
	if !ok {
		return
	}

	err := c.pipelineService.CreateLogs(ctx, buildS3LogsParam(pipelineID, executionID, stageID, stepID), ctx.Request.Body)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionLogCreateError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.ExecutionLogCreateSuccess, nil)
}

// @Summary 流式获取流水线步骤日志（纯文本/流）
// @Description 优先从文件存储读取已完成的日志，若无文件日志则以 SSE 方式流式返回实时日志流，输出格式为纯文本 [时间戳] 消息
// @Tags executions
// @Produce text/event-stream
// @Param pipelineID path string true "流水线ID"
// @Param executionID path string true "执行ID"
// @Param stageID path string true "阶段ID"
// @Param stepID path string true "步骤ID"
// @Success 200 {string} string "SSE 日志流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/executions/{executionID}/logstream/{stageID}/{stepID} [get]
func (c *Controller) GetLogStream(ctx *gin.Context) {
	pipelineID, executionID, stageID, stepID, ok := parseLogPathParams(ctx)
	if !ok {
		return
	}

	f, ok := setupSSE(ctx)
	if !ok {
		return
	}

	w := ctx.Writer

	// 优先尝试从文件存储读取已完成的日志
	rc, fileErr := c.pipelineService.GetLogs(ctx, buildS3LogsParam(pipelineID, executionID, stageID, stepID))
	if fileErr == nil {
		defer rc.Close()
		data, _ := io.ReadAll(rc)
		// 尝试 JSON 数组格式 [{...},{...}]
		var lines []*livelog.Line
		if json.Unmarshal(data, &lines) == nil {
			for _, line := range lines {
				writeLineAsText(w, line)
				f.Flush()
			}
		} else {
			// 回退到按行扫描（JSON Lines 或纯文本）
			scanner := bufio.NewScanner(bytes.NewReader(data))
			for scanner.Scan() {
				raw := scanner.Bytes()
				var line livelog.Line
				if json.Unmarshal(raw, &line) == nil {
					writeLineAsText(w, &line)
				} else {
					fmt.Fprintf(w, "data: %s\n\n", raw)
				}
				f.Flush()
			}
		}
		io.WriteString(w, "event: error\ndata: eof\n\n")
		f.Flush()
		return
	}

	// 文件日志不存在，回退到实时日志流
	c.tailLiveStream(ctx, stepID, f, writeLineAsText)
}
