// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package deploy

import (
	"context"

	repo_model "gitea.dev/models/repo"
	user_model "gitea.dev/models/user"
	"gitea.dev/modules/log"
	"gitea.dev/modules/repository"
	"gitea.dev/modules/setting"
	notify_service "gitea.dev/services/notify"
)

func init() {
	notify_service.RegisterNotifier(new(deployNotifier))
}

type deployNotifier struct {
	notify_service.NullNotifier
}

// PushCommits deploys pushes to the branch of every environment that deploys on push
func (*deployNotifier) PushCommits(ctx context.Context, pusher *user_model.User, repo *repo_model.Repository, opts *repository.PushUpdateOptions, _ *repository.PushCommits) {
	if !setting.Jenkins.Enabled || opts.IsDelRef() || !opts.RefFullName.IsBranch() {
		return
	}
	envs, err := GetEnvironments(ctx, repo.ID)
	if err != nil {
		log.Error("PushCommits: GetEnvironments for %s: %v", repo.FullName(), err)
		return
	}
	for _, env := range envs {
		if env.Trigger == "push" && env.Branch == opts.RefFullName.BranchName() {
			if err := Deploy(ctx, repo, pusher, env, opts.NewCommitID); err != nil {
				log.Error("PushCommits: deploy %s to %s: %v", repo.FullName(), env.Name, err)
			}
		}
	}
}
