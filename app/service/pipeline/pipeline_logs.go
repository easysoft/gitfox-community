package pipeline

import (
	"context"
	"io"
	"time"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/pkg/livelog"
)

func (s *Service) CreateLogs(ctx context.Context, param *pipelineModel.S3Logs, r io.Reader) error {
	// contentStorage 已配置 pipeline 前缀，Key() 返回 /<pipelineID>/<executionID>/<stageID>/<stepID>.log
	return s.contentStorage.Save(ctx, param.Key(), r, -1)
}

func (s *Service) GetLogs(ctx context.Context, param *pipelineModel.S3Logs) (io.ReadCloser, error) {
	return s.contentStorage.Open(ctx, param.Key())
}

// CreateLogStream 创建日志流
func (s *Service) CreateLogStream(ctx context.Context, stepID int64) error {
	return s.logStream.Create(ctx, stepID)
}

// DeleteLogStream 删除日志流
func (s *Service) DeleteLogStream(ctx context.Context, stepID int64) error {
	return s.logStream.Delete(ctx, stepID)
}

// WriteLogStream 写入日志流
func (s *Service) WriteLogStream(ctx context.Context, stepID int64, line *livelog.Line) error {
	return s.logStream.Write(ctx, stepID, line)
}

// TailLogStream 监听日志流。若流尚未创建，会轮询等待直到流可用或 ctx 取消。
// 始终返回非 nil 的 channel，调用方可直接 select 使用。
func (s *Service) TailLogStream(ctx context.Context, stepID int64) (<-chan *livelog.Line, <-chan error) {
	// 流已存在，直接返回
	linec, errc := s.logStream.Tail(ctx, stepID)
	if errc != nil {
		return linec, errc
	}

	// 流不存在，启动后台 goroutine 轮询等待
	lineCh := make(chan *livelog.Line, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(lineCh)
		defer close(errCh)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// 等待流创建
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-ticker.C:
				lc, ec := s.logStream.Tail(ctx, stepID)
				if ec == nil {
					continue // 流仍未创建，继续等待
				}
				// 流已创建，转发数据
				s.forwardStream(ctx, lc, ec, lineCh, errCh)
				return
			}
		}
	}()

	return lineCh, errCh
}

// forwardStream 将底层 logStream 的 channel 数据转发到代理 channel。
func (s *Service) forwardStream(
	ctx context.Context,
	lc <-chan *livelog.Line,
	ec <-chan error,
	lineCh chan<- *livelog.Line,
	errCh chan<- error,
) {
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		case err, ok := <-ec:
			if ok {
				errCh <- err
			}
			return
		case line, ok := <-lc:
			if !ok {
				return
			}
			select {
			case lineCh <- line:
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}
	}
}

// InfoLogStream 获取日志流信息
func (s *Service) InfoLogStream(ctx context.Context) *livelog.LogStreamInfo {
	return s.logStream.Info(ctx)
}
