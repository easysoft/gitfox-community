package git

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

func (ctl *GitController) PreReceiveClone(ctx *gin.Context) {
	resp.OK(ctx, i18n.GitGitHookPreReceiveSuccess, nil)
}

func (ctl *GitController) PostReceiveClone(ctx *gin.Context) {
	resp.OK(ctx, i18n.GitGitHookPostReceiveSuccess, nil)
}

func (ctl *GitController) UpdateClone(ctx *gin.Context) {
	resp.OK(ctx, i18n.TagDeleteSuccess, nil)
}
