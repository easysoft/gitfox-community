package artifact

import (
	"fmt"
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 创建分组
// @Description 创建 group 节点。artifactID 必传且必须大于 0
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.CreateGroupReq true "请求参数(artifactID 必填；names 为 '.' 分隔的分组路径；format 必须是支持的格式枚举)"
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

	// 检查视图是否存在
	if err := c.artSvc.ViewExists(ctx, params.ArtifactID); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	names := params.ParseNames()
	format := model.ArtifactFormat(params.Format)

	// 检查分组路径是否已存在（防止重复创建）
	exists, err := c.artSvc.GroupPathExists(ctx, params.ArtifactID, names, format)
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

	ge, err := c.artSvc.CreateGroup(ctx, params.ArtifactID, names, format)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactCreateGroupSuccess, ge)
}

// @Summary 查询分组文件列表
// @Description 列出指定分组的制品文件列表
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param groupId path int true "分组ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/groups/{groupId}/assets [get]
func (c *Controller) ListGroupAssets(ctx *gin.Context) {
	groupId, err := req.PathParamAsPositiveInt64(ctx, "groupId")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	assets, err := c.artSvc.ListGroupAssets(ctx, groupId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, assets)
}

// @Summary 归档 generic 分组
// @Description 将 generic 分组资产打包为 zip 下载
// @Tags artifacts
// @Produce application/zip
// @Param groupId path int true "分组ID"
// @Param spaceId query int true "空间ID"
// @Success 200 {file} file "ZIP 文件流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/groups/{groupId}/archive [get]
func (c *Controller) ArchiveGroup(ctx *gin.Context) {
	groupID, err := req.PathParamAsPositiveInt64(ctx, "groupId")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceId")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}
	session, _ := req.GetAuthSession(ctx)
	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoId"); rID > 0 {
		repoID = rID
	}
	archive, err := c.artSvc.PrepareArchiveGroup(ctx, session, spaceID, repoID, groupID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", "attachment; filename="+archive.Name())
	ctx.Status(http.StatusOK)
	if err := c.artSvc.ArchiveGroup(ctx, archive, ctx.Writer); err != nil {
		resp.InternalError(ctx, err)
		return
	}
}
