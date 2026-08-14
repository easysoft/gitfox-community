package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

func (c *ScanController) Rules(ctx *gin.Context) {
	result, err := c.aicodereviewSvc.GetAIReviewRules(ctx.Request.Context())
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	resp.ListSuccess(ctx, result, nil, i18n.AICodeReviewRuleListSuccess)
}
