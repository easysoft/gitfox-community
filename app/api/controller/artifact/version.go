package artifact

// @Summary 查询版本列表
// @Description 列出指定制品包的版本列表。artifactID > 0 时优先按指定视图查询，否则按默认 view 逻辑查询
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.ArtifactVersionListReq true "查询条件"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/versions/list [post]
// func (c *Controller) ListVersions(ctx *gin.Context) {
// 	var reqBody model.ArtifactVersionListReq
// 	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
// 		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
// 		resp.BadRequest(ctx, err)
// 		return
// 	}

// 	if err := reqBody.Validate(); err != nil {
// 		resp.BadRequest(ctx, err)
// 		return
// 	}
// 	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody.SpaceRepoReq)
// 	if err != nil {
// 		resp.BadRequest(ctx, err)
// 		return
// 	}

// 	versions, err := c.artSvc.ListVersions(ctx, view.ArtifactID, reqBody.PackageName, view.ArtifactName, view.ArtifactFormat, model.SearchVersionOption{})
// 	if err != nil {
// 		resp.InternalError(ctx, err)
// 		return
// 	}
// 	resp.OK(ctx, i18n.ArtifactListSuccess, versions)
// }
