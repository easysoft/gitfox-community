package resp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Stream[T any] interface {
	Next() (T, error)
}

// JSONArrayDynamic outputs an JSON array whose elements are streamed from a channel.
// Due to the dynamic nature (unknown number of elements) the function will use
// chunked transfer encoding for large files.
func JSONArrayDynamic[T comparable](ctx *gin.Context, w http.ResponseWriter, stream Stream[T]) {
	count := 0
	enc := json.NewEncoder(w)
	log := logger.GetLoggerFromGin(ctx)

	for {
		data, err := stream.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			// based on discussion unrelated histories error should return
			// StatusOK on diff.
			if uErr := api.AsUnrelatedHistoriesError(err); uErr != nil {
				InternalError(ctx, uErr)
				return
			}
			// User canceled the request - no need to do anything
			if errors.Is(err, context.Canceled) {
				return
			}

			if count == 0 {
				InternalError(ctx, err)
				return
			}

			// Array data has been already streamed, it's too late for the output - so just log and quit.
			// log.Warn().Err(err).Msgf("Failed to write JSON array response body")
			log.Warnf("JSON array response body has been already streamed, it's too late for the output.")
			// close array
			_, _ = w.Write([]byte{']'})
			return
		}

		if count == 0 {
			setCommonHeaders(w)
			_, _ = w.Write([]byte{'['})
		} else {
			_, _ = w.Write([]byte{','})
		}

		count++

		_ = enc.Encode(data)
	}

	if count == 0 {
		setCommonHeaders(w)
		_, _ = w.Write([]byte{'['})
	}

	_, _ = w.Write([]byte{']'})
}
