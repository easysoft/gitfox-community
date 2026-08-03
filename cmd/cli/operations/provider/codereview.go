package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/review"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	reviewSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/codereview"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	pipelineStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

// CodeReviewWireSet provides a wire set for code review service.
var CodeReviewWireSet = wire.NewSet(
	review.NewReviewController,
	ProvideCodeReviewService,
	wire.Bind(new(reviewSvc.Interface), new(*reviewSvc.CodeReviewService)),
)

// ProvideCodeReviewService creates a new code review service.
func ProvideCodeReviewService(
	config *config.Config,
	pipelineStore pipelineStore.Interface,
	pipelineSvc pipelineSvc.Interface,
	triggerer triggerer.Triggerer,
) *reviewSvc.CodeReviewService {
	return reviewSvc.NewCodeReviewService(
		config,
		pipelineStore,
		pipelineSvc,
		triggerer,
	)
}
