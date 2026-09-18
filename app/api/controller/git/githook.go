package git

import (
	"fmt"

	githookModel "github.com/easysoft/gitfox-community/app/model/githook"
	"github.com/easysoft/gitfox-community/pkg/git/hook"

	"github.com/gin-gonic/gin"
)

// PreReceive handles the pre-receive git hook
func (ctl *GitController) PreReceive(ctx *gin.Context) {

	// Parse request body
	var input githookModel.GithookPreReceiveInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		errMsg := fmt.Sprintf("failed to parse pre-receive input: %v", err)
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    &errMsg,
		})
		return
	}

	// Call the actual pre-receive logic
	output, err := ctl.githookCtrl.PreReceive(ctx, ctl.git, input)
	if output.Error != nil {
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    output.Error,
		})
		return
	}
	if err != nil {
		errMsg := fmt.Sprintf("pre-receive hook failed: %v", err)
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    &errMsg,
		})
		return
	}

	ctx.JSON(200, hook.Output{})
}

// PostReceive handles the post-receive git hook
func (ctl *GitController) PostReceive(ctx *gin.Context) {

	// Parse request body
	var input githookModel.GithookPostReceiveInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		errMsg := fmt.Sprintf("failed to parse post-receive input: %v", err)
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    &errMsg,
		})
		return
	}

	// Call the actual post-receive logic
	output, err := ctl.githookCtrl.PostReceive(ctx, ctl.git, input)
	if output.Error != nil {
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    output.Error,
		})
		return
	}
	if err != nil {
		errMsg := fmt.Sprintf("post-receive hook failed: %v", err)
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    &errMsg,
		})
		return
	}

	ctx.JSON(200, hook.Output{})
}

// Update handles the update git hook
func (ctl *GitController) Update(ctx *gin.Context) {

	// Parse request body
	var input githookModel.GithookUpdateInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		errMsg := fmt.Sprintf("failed to parse update input: %v", err)
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    &errMsg,
		})
		return
	}

	// Call the actual update logic
	output, err := ctl.githookCtrl.Update(ctx, ctl.git, input)
	if output.Error != nil {
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    output.Error,
		})
		return
	}
	if err != nil {
		errMsg := fmt.Sprintf("update hook failed: %v", err)
		ctx.JSON(400, hook.Output{
			Messages: nil,
			Error:    &errMsg,
		})
		return
	}

	ctx.JSON(200, hook.Output{})
}
