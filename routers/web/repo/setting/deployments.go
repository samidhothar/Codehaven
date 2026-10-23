// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"errors"
	"net/http"

	"gitea.dev/models/db"
	git_model "gitea.dev/models/git"
	"gitea.dev/modules/optional"
	"gitea.dev/modules/setting"
	"gitea.dev/modules/templates"
	"gitea.dev/modules/util"
	"gitea.dev/services/context"
	deploy_service "gitea.dev/services/deploy"
)

const tplDeployments templates.TplName = "repo/settings/deployments"

func Deployments(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.settings.deployments")
	ctx.Data["PageIsSettingsDeployments"] = true

	envs, err := deploy_service.GetEnvironments(ctx, ctx.Repo.Repository.ID)
	if err != nil {
		ctx.ServerError("GetEnvironments", err)
		return
	}
	branches, err := git_model.FindBranchNames(ctx, git_model.FindBranchOptions{
		RepoID:          ctx.Repo.Repository.ID,
		IsDeletedBranch: optional.Some(false),
		ListOptions:     db.ListOptionsAll,
	})
	if err != nil {
		ctx.ServerError("FindBranchNames", err)
		return
	}
	ctx.Data["DeployEnabled"] = deploy_service.Enabled()
	ctx.Data["DeployEnvironments"] = envs
	ctx.Data["DeployEnvironmentNames"] = deploy_service.EnvironmentNames
	ctx.Data["DeployTriggers"] = deploy_service.Triggers
	ctx.Data["DeployPresets"] = deploy_service.Presets
	ctx.Data["DeployBranches"] = branches
	ctx.Data["DeploySSHCredentialsID"] = setting.Jenkins.SSHCredentialsID
	ctx.Data["NewDeployEnvironment"] = &deploy_service.Environment{Trigger: "push", Branch: ctx.Repo.Repository.DefaultBranch, Preset: "static", Port: 22}
	ctx.HTML(http.StatusOK, tplDeployments)
}

func DeploymentsPost(ctx *context.Context) {
	var err error
	switch ctx.FormString("action") {
	case "save":
		err = deploy_service.SaveEnvironment(ctx, ctx.Repo.Repository.ID, &deploy_service.Environment{
			Name:    ctx.FormString("name"),
			Trigger: ctx.FormString("trigger"),
			Branch:  ctx.FormTrim("branch"),
			Preset:  ctx.FormString("preset"),
			Host:    ctx.FormTrim("host"),
			Port:    ctx.FormInt("port"),
			User:    ctx.FormTrim("user"),
			Path:    ctx.FormTrim("path"),
			Command: ctx.FormTrim("command"),
		})
	case "delete":
		err = deploy_service.DeleteEnvironment(ctx, ctx.Repo.Repository.ID, ctx.FormString("name"))
	default:
		ctx.HTTPError(http.StatusBadRequest)
		return
	}

	switch {
	case err == nil:
		ctx.Flash.Success(ctx.Tr("repo.settings.deployments.saved"))
		ctx.JSONRedirect(ctx.Repo.RepoLink + "/settings/deployments")
	case errors.Is(err, util.ErrInvalidArgument):
		ctx.JSONError(ctx.Tr("repo.settings.deployments.invalid", err.Error()))
	default:
		ctx.ServerError("DeploymentsPost", err)
	}
}
