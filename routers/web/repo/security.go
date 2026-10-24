// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"
	"slices"

	"gitea.dev/modules/log"
	"gitea.dev/modules/templates"
	"gitea.dev/services/context"
	"gitea.dev/services/securityscan"
)

const tplSecurity templates.TplName = "repo/security"

// Security shows the findings of the latest Trivy scan uploaded by a workflow run
func Security(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("repo.security")
	ctx.Data["PageIsSecurity"] = true
	ctx.Data["ArtifactName"] = securityscan.ArtifactName
	ctx.Data["Kinds"] = securityscan.Kinds
	ctx.Data["Severities"] = securityscan.Severities

	scan, err := securityscan.FindLatest(ctx, ctx.Repo.Repository)
	if err != nil {
		// a broken or expired artifact should not hide the page and its setup help
		log.Warn("Security: %v", err)
		ctx.Data["ReportBroken"] = true
	}
	if scan != nil {
		kind, severity := ctx.FormString("kind"), ctx.FormString("severity")
		if !slices.Contains(securityscan.Kinds, kind) {
			kind = ""
		}
		if !slices.Contains(securityscan.Severities, severity) {
			severity = ""
		}
		ctx.Data["Scan"] = scan
		ctx.Data["Findings"] = scan.Report.Filter(kind, severity)
		ctx.Data["Kind"] = kind
		ctx.Data["Severity"] = severity
	}
	ctx.HTML(http.StatusOK, tplSecurity)
}
