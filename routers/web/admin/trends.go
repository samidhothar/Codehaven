// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"
	"time"

	"gitea.dev/modules/setting"
	"gitea.dev/modules/templates"
	"gitea.dev/services/context"
	"gitea.dev/services/insights"
)

const tplTrends templates.TplName = "admin/trends"

// MonitorTrends shows weekly activity across every repository of the server
func MonitorTrends(ctx *context.Context) {
	weeks := insights.ParseWeeks(ctx.FormInt("weeks"))
	trends, err := insights.Collect(ctx, nil, time.Now(), weeks)
	if err != nil {
		ctx.ServerError("MonitorTrends", err)
		return
	}
	ctx.Data["Title"] = ctx.Tr("insights.trends")
	ctx.Data["PageIsAdminMonitorTrends"] = true
	ctx.Data["Trends"] = trends
	ctx.Data["TrendsWeeks"] = weeks
	ctx.Data["TrendsPeriods"] = insights.PeriodWeeks
	ctx.Data["TrendsLink"] = setting.AppSubURL + "/-/admin/monitor/trends"
	ctx.HTML(http.StatusOK, tplTrends)
}
