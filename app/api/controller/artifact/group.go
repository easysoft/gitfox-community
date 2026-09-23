// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"fmt"

	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 创建分组
// @Description 创建 group 节点。artifactID 必传且必须大于 0。group 的 format 自动从 artifactID 对应视图中反查，前端不需要传 format。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.CreateGroupReq true "请求参数(artifactID 必填；names 为 '.' 分隔的分组路径；format 不传，由服务端从 artifact 反查)"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/groups [post]
func (c *Controller) CreateGroup(ctx *gin.Context) {
	var params model.CreateGroupReq
	if err := ctx.ShouldBindBodyWithJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 业务规则校验（artifactID > 0、format 枚举、names 合法、container 不嵌套）
	if err := params.Validate(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	// 反查 lib，确保 artifactID 存在且 format 合法
	lib, err := c.artSvc.ResolveLib(ctx.Request.Context(), &model.SpaceRepoReq{ArtifactID: params.ArtifactID})
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	// 用 lib.Format 覆盖前端可能传入的 format，保证 group.format 与 artifact 绑定一致
	params.Format = string(lib.Format)

	names := params.ParseNames()
	format := lib.Format

	// 检查分组路径是否已存在（仅活的 group 算重复，软删的允许重建）
	exists, err := c.artSvc.ActiveGroupPathExists(ctx.Request.Context(), params.ArtifactID, names, format)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	if exists {
		resp.BadRequest(ctx, errors.WithI18nKey(
			fmt.Errorf("group path %q already exists under artifactID=%d format=%s", params.Names, params.ArtifactID, params.Format),
			i18n.ArtifactGroupDuplicate,
		))
		return
	}

	ge, err := c.artSvc.CreateGroup(ctx.Request.Context(), params.ArtifactID, names, format)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactCreateGroupSuccess, ge)
}

// @Summary 归档 file 分组
// @Description 将 file 分组资产打包为 zip 下载
// @Tags artifacts
// @Produce application/zip
// @Param groupID path int true "分组ID"
// @Param spaceID query int true "空间ID"
// @Success 200 {file} file "ZIP 文件流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/groups/{groupID}/archive [get]
// func (c *Controller) ArchiveGroup(ctx *gin.Context) {
// 	groupID, err := req.PathParamAsPositiveInt64(ctx, "groupID")
// 	if err != nil {
// 		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
// 		resp.BadRequest(ctx, err)
// 		return
// 	}
// 	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceID")
// 	if err != nil {
// 		err = errors.WithI18nKey(err, i18n.ArtifactQueryParamError)
// 		resp.BadRequest(ctx, err)
// 		return
// 	}
// 	session, _ := req.GetAuthSession(ctx)
// 	var repoID int64
// 	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoID"); rID > 0 {
// 		repoID = rID
// 	}
// 	archive, err := c.artSvc.PrepareArchiveGroup(ctx, session, spaceID, repoID, groupID)
// 	if err != nil {
// 		resp.InternalError(ctx, err)
// 		return
// 	}
// 	ctx.Header("Content-Type", "application/zip")
// 	ctx.Header("Content-Disposition", "attachment; filename="+archive.Name())
// 	ctx.Status(http.StatusOK)
// 	if err := c.artSvc.ArchiveGroup(ctx, archive, ctx.Writer); err != nil {
// 		resp.InternalError(ctx, err)
// 		return
// 	}
// }
