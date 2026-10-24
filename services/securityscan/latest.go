// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package securityscan

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"

	actions_model "gitea.dev/models/actions"
	"gitea.dev/models/db"
	repo_model "gitea.dev/models/repo"
	"gitea.dev/modules/git"
	"gitea.dev/modules/storage"

	"xorm.io/builder"
)

// ArtifactName is the artifact a workflow uploads the Trivy JSON report as
const ArtifactName = "trivy-results"

const maxReportSize = 32 * 1024 * 1024

type Scan struct {
	Run       *actions_model.ActionRun
	CommitSHA string
	Report    *Report
}

// FindLatest returns the report of the newest successful run on the default branch, or nil when there is none
func FindLatest(ctx context.Context, repo *repo_model.Repository) (*Scan, error) {
	runs := builder.Select("id").From("action_run").Where(builder.Eq{
		"repo_id": repo.ID,
		"ref":     git.BranchPrefix + repo.DefaultBranch,
		"status":  actions_model.StatusSuccess,
	})
	artifact := &actions_model.ActionArtifact{}
	has, err := db.GetEngine(ctx).Where(builder.Eq{
		"repo_id":       repo.ID,
		"artifact_name": ArtifactName,
		"status":        actions_model.ArtifactStatusUploadConfirmed,
	}.And(builder.In("run_id", runs))).Desc("id").Get(artifact)
	if err != nil {
		return nil, fmt.Errorf("FindLatest: %w", err)
	} else if !has {
		return nil, nil
	}

	run, err := actions_model.GetRunByRepoAndID(ctx, repo.ID, artifact.RunID)
	if err != nil {
		return nil, fmt.Errorf("FindLatest: %w", err)
	}
	run.Repo = repo
	report, err := readArtifact(artifact)
	if err != nil {
		return nil, fmt.Errorf("FindLatest: %w", err)
	}
	return &Scan{Run: run, CommitSHA: artifact.CommitSHA, Report: report}, nil
}

func readArtifact(artifact *actions_model.ActionArtifact) (*Report, error) {
	if artifact.FileSize > maxReportSize {
		return nil, fmt.Errorf("readArtifact: report is larger than %d bytes", maxReportSize)
	}
	obj, err := storage.ActionsArtifacts.Open(artifact.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("readArtifact: %w", err)
	}
	defer obj.Close()

	var content io.Reader = obj
	switch artifact.ContentEncodingOrType {
	case actions_model.ContentEncodingV3Gzip:
		gz, err := gzip.NewReader(obj)
		if err != nil {
			return nil, fmt.Errorf("readArtifact: %w", err)
		}
		defer gz.Close()
		content = gz
	case actions_model.ContentTypeZip:
		data, err := io.ReadAll(io.LimitReader(obj, maxReportSize))
		if err != nil {
			return nil, fmt.Errorf("readArtifact: %w", err)
		}
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, fmt.Errorf("readArtifact: %w", err)
		}
		f := findJSONFile(zr)
		if f == nil {
			return nil, fmt.Errorf("readArtifact: no JSON file in artifact %q", artifact.ArtifactName)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("readArtifact: %w", err)
		}
		defer rc.Close()
		content = rc
	}
	return Parse(io.LimitReader(content, maxReportSize)) // also bounds decompressed content
}

func findJSONFile(zr *zip.Reader) *zip.File {
	for _, f := range zr.File {
		if !f.Mode().IsDir() && strings.HasSuffix(strings.ToLower(f.Name), ".json") {
			return f
		}
	}
	return nil
}
