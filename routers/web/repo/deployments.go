// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"errors"
	"net/http"

	"gitea.dev/modules/templates"
	"gitea.dev/modules/util"
	"gitea.dev/services/context"
	deploy_service "gitea.dev/services/deploy"
)

const tplDeploymentsMenu templates.TplName = "repo/deployments_menu"

// DeploymentsMenu renders the content of the repository header's deploy dropdown
func DeploymentsMenu(ctx *context.Context) {
	envs, err := deploy_service.GetEnvironments(ctx, ctx.Repo.Repository.ID)
	if err != nil {
		ctx.ServerError("GetEnvironments", err)
		return
	}
	ctx.Data["DeployEnvironments"] = envs
	ctx.HTML(http.StatusOK, tplDeploymentsMenu)
}

// DeploymentRun deploys the head of the environment's branch
func DeploymentRun(ctx *context.Context) {
	env, err := deploy_service.GetEnvironment(ctx, ctx.Repo.Repository.ID, ctx.PathParam("env"))
	if errors.Is(err, util.ErrNotExist) {
		ctx.JSONError(ctx.Tr("error.not_found"))
		return
	} else if err != nil {
		ctx.ServerError("GetEnvironment", err)
		return
	}
	commitID, err := ctx.Repo.GitRepo.GetBranchCommitID(ctx, env.Branch)
	if err != nil {
		ctx.JSONError(ctx.Tr("repo.deployments.branch_missing", env.Branch))
		return
	}
	if err := deploy_service.Deploy(ctx, ctx.Repo.Repository, ctx.Doer, env, commitID); err != nil {
		if errors.Is(err, util.ErrInvalidArgument) {
			ctx.JSONError(err.Error())
			return
		}
		ctx.JSONError(ctx.Tr("repo.deployments.start_failed", err.Error()))
		return
	}
	ctx.Flash.Success(ctx.Tr("repo.deployments.started", env.Name, env.Branch))
	ctx.JSONRedirect(ctx.Repo.RepoLink)
}
