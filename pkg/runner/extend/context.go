package extend

import (
	"context"

	schema "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"
)

type stageContext struct {
	stage    *schema.Stage
	pipeline *schema.Pipeline
}

type contextKey struct{}

func WithContext(ctx context.Context, stage *schema.Stage, pipeline *schema.Pipeline) context.Context {
	return context.WithValue(ctx, contextKey{}, &stageContext{
		stage:    stage,
		pipeline: pipeline,
	})
}

func FromContext(ctx context.Context) *stageContext {
	obj, ok := ctx.Value(contextKey{}).(*stageContext)
	if !ok {
		return nil
	}
	return obj
}
