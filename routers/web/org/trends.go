// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package org

import (
	"net/http"
	"time"

	repo_model "gitea.dev/models/repo"
	"gitea.dev/models/unit"
	"gitea.dev/modules/templates"
	shared_user "gitea.dev/routers/web/shared/user"
	"gitea.dev/services/context"
	"gitea.dev/services/insights"

	"xorm.io/builder"
)

const tplTrends templates.TplName = "org/trends"

// Trends shows weekly activity across the organization's repositories that the viewer can access
func Trends(ctx *context.Context) {
	weeks := insights.ParseWeeks(ctx.FormInt("weeks"))
	repoCond := builder.And(
		builder.Eq{"`repository`.owner_id": ctx.Org.Organization.ID},
		repo_model.AccessibleRepositoryCondition(ctx.Doer, unit.TypeInvalid),
	)
	trends, err := insights.Collect(ctx, repoCond, time.Now(), weeks)
	if err != nil {
		ctx.ServerError("Trends", err)
		return
	}
	if _, err := shared_user.RenderUserOrgHeader(ctx); err != nil {
		ctx.ServerError("RenderUserOrgHeader", err)
		return
	}
	ctx.Data["Title"] = ctx.Tr("insights.trends")
	ctx.Data["PageIsOrgTrends"] = true
	ctx.Data["Trends"] = trends
	ctx.Data["TrendsWeeks"] = weeks
	ctx.Data["TrendsPeriods"] = insights.PeriodWeeks
	ctx.Data["TrendsLink"] = ctx.Org.OrgLink + "/trends"
	ctx.HTML(http.StatusOK, tplTrends)
}
