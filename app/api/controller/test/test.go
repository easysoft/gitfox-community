package test

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// TestController 处理与测试相关的 HTTP 请求
type TestController struct {
}

// NewTestController 创建测试处理器实例
func NewTestController() *TestController {
	return &TestController{}
}

func (ctl *TestController) Test(c *gin.Context) {
	resp.Error(c, errors.WithI18nKey(errors.New("出错误了？"), i18n.ErrorDefault), "test error")
}
