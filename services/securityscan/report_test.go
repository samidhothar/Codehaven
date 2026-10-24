// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package securityscan

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	report, err := Parse(strings.NewReader(`{
		"SchemaVersion": 2, "ArtifactType": "filesystem",
		"Results": [
			{"Target": "package-lock.json", "Class": "lang-pkgs", "Vulnerabilities": [
				{"VulnerabilityID": "CVE-1", "PkgName": "lodash", "InstalledVersion": "4.17.15", "FixedVersion": "4.17.21", "Severity": "HIGH", "Title": "Prototype pollution"},
				{"VulnerabilityID": "CVE-2", "PkgName": "minimist", "Severity": "critical"}
			]},
			{"Target": "Dockerfile", "Class": "config", "Misconfigurations": [
				{"ID": "DS002", "Title": "Runs as root", "Severity": "HIGH", "Status": "FAIL", "CauseMetadata": {"StartLine": 4}},
				{"ID": "DS001", "Severity": "LOW", "Status": "PASS"}
			]},
			{"Target": ".env", "Class": "secret", "Secrets": [
				{"RuleID": "aws-access-key-id", "Title": "AWS Access Key ID", "Severity": "SEVERE", "StartLine": 2, "Match": "AKIA****"}
			]}
		]
	}`))
	require.NoError(t, err)

	ids := make([]string, len(report.Findings))
	for i, f := range report.Findings {
		ids[i] = f.ID
	}
	assert.Equal(t, []string{"CVE-2", "CVE-1", "DS002", "aws-access-key-id"}, ids) // severity first, then kind
	assert.Equal(t, map[string]int{"CRITICAL": 1, "HIGH": 2, "UNKNOWN": 1}, report.SeverityCounts)
	assert.Equal(t, map[string]int{KindVulnerability: 2, KindMisconfig: 1, KindSecret: 1}, report.KindCounts)

	assert.Equal(t, "CVE-2", report.Findings[0].Title) // falls back to the ID
	assert.Equal(t, &Finding{Kind: KindMisconfig, ID: "DS002", Title: "Runs as root", Severity: "HIGH", Target: "Dockerfile", IsFile: true, Line: 4}, report.Findings[2])

	assert.Len(t, report.Filter(KindVulnerability, ""), 2)
	assert.Len(t, report.Filter("", "HIGH"), 2)
	assert.Len(t, report.Filter(KindVulnerability, "HIGH"), 1)
	assert.Len(t, report.Findings, 4, "filtering must not modify the report")

	image, err := Parse(strings.NewReader(`{"ArtifactType": "container_image", "Results": [{"Target": "alpine 3.22", "Class": "os-pkgs", "Vulnerabilities": [{"VulnerabilityID": "CVE-3"}]}]}`))
	require.NoError(t, err)
	assert.False(t, image.Findings[0].IsFile)

	_, err = Parse(strings.NewReader("not json"))
	assert.Error(t, err)
}
