// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package extend

import (
	"context"

	schema "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
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
