package artifact

// @Summary 容量统计
// @Description 返回非容器格式制品容量统计
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/capacity/all [get]
// func (c *Controller) StatisticCapacity(ctx *gin.Context) {
// 	data, err := c.artSvc.StatisticCapacity(ctx)
// 	if err != nil {
// 		resp.InternalError(ctx, err)
// 		return
// 	}
// 	resp.OK(ctx, i18n.ArtifactStatisticSuccess, data)
// }

// @Summary 容器容量统计
// @Description 返回 container 格式制品容量统计
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/capacity/container [get]
// func (c *Controller) StatisticContainerCapacity(ctx *gin.Context) {
// 	data, err := c.artSvc.StatisticContainerCapacity(ctx)
// 	if err != nil {
// 		resp.InternalError(ctx, err)
// 		return
// 	}
// 	resp.OK(ctx, i18n.ArtifactStatisticSuccess, data)
// }

// @Summary 制品汇总统计
// @Description 返回制品资产汇总统计
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/statistic [get]
// func (c *Controller) StatisticSummary(ctx *gin.Context) {
// 	data, err := c.artSvc.StatisticSummary(ctx)
// 	if err != nil {
// 		resp.InternalError(ctx, err)
// 		return
// 	}
// 	resp.OK(ctx, i18n.ArtifactStatisticSuccess, data)
// }
