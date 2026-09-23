// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package resp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

func StreamSSE(
	ctx context.Context,
	w http.ResponseWriter,
	chStop <-chan struct{},
	chEvents <-chan *sse.Event,
	chErr <-chan error,
) {
	logger := logger.GetLogger(ctx)

	flusher, ok := w.(http.Flusher)
	if !ok {
		//UserError(ctx, w, usererror.ErrResponseNotFlushable)
		logger.Errorf("failed to build SSE stream")
		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	h.Set("Access-Control-Allow-Origin", "*")

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	stream := sseStream{
		enc:     enc,
		writer:  w,
		flusher: flusher,
	}

	const (
		pingInterval = 30 * time.Second
		tailMaxTime  = 2 * time.Hour
	)

	ctx, ctxCancel := context.WithTimeout(ctx, tailMaxTime)
	defer ctxCancel()

	if err := stream.ping(); err != nil {
		logger.Errorf("failed to send initial ping, err:%v", err)
		return
	}

	defer func() {
		if err := stream.close(); err != nil {
			logger.Errorf("failed to close SSE stream, err:%v", err)
		}
	}()

	pingTimer := time.NewTimer(pingInterval)
	defer pingTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debugf("stream SSE request context done, err:%v", ctx.Err())
			return

		case <-chStop:
			logger.Debugf("app shutdown")
			return

		case err := <-chErr:
			logger.Debugf("received error from SSE stream, err:%v", err)
			return

		case <-pingTimer.C:
			if err := stream.ping(); err != nil {
				logger.Errorf("failed to send SSE ping, err:%v", err)
				return
			}

		case event, canProduce := <-chEvents:
			if !canProduce {
				logger.Debugf("events channel is drained and closed.")
				return
			}
			if err := stream.event(event); err != nil {
				logger.Errorf("failed to send SSE event: %s, err:%v", event.Type, err)
				return
			}
		}

		pingTimer.Stop() // stop timer

		select {
		case <-pingTimer.C: // drain channel
		default:
		}

		pingTimer.Reset(pingInterval) // reset timer
	}
}

type sseStream struct {
	enc     *json.Encoder
	writer  io.Writer
	flusher http.Flusher
}

func (r sseStream) event(event *sse.Event) error {
	_, err := io.WriteString(r.writer, fmt.Sprintf("event: %s\n", event.Type))
	if err != nil {
		return fmt.Errorf("failed to send event header: %w", err)
	}

	_, err = io.WriteString(r.writer, "data: ")
	if err != nil {
		return fmt.Errorf("failed to send data header: %w", err)
	}

	err = r.enc.Encode(event.Data)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	// NOTE: enc.Encode is ending the data with a new line, only add one more
	// Source: https://cs.opensource.google/go/go/+/refs/tags/go1.21.1:src/encoding/json/stream.go;l=220
	_, err = r.writer.Write([]byte{'\n'})
	if err != nil {
		return fmt.Errorf("failed to send end of message: %w", err)
	}

	r.flusher.Flush()
	return nil
}

func (r sseStream) close() error {
	_, err := io.WriteString(r.writer, "event: error\ndata: eof\n\n")
	if err != nil {
		return fmt.Errorf("failed to send EOF: %w", err)
	}
	r.flusher.Flush()
	return nil
}

func (r sseStream) ping() error {
	_, err := io.WriteString(r.writer, ": ping\n\n")
	if err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}
	r.flusher.Flush()
	return nil
}
