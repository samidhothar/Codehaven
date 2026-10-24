// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package securityscan

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"

	"gitea.dev/modules/json"
)

const (
	KindVulnerability = "vulnerability"
	KindSecret        = "secret"
	KindMisconfig     = "misconfig"
)

var (
	Kinds      = []string{KindVulnerability, KindSecret, KindMisconfig}
	Severities = []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"}
)

type Finding struct {
	Kind      string
	ID        string
	Title     string
	Severity  string
	Target    string
	IsFile    bool // Target is a path in the repository and can be linked
	Line      int
	Package   string
	Installed string
	Fixed     string
	URL       string
}

type Report struct {
	Findings       []*Finding
	SeverityCounts map[string]int
	KindCounts     map[string]int
}

// trivyReport is the subset of Trivy's "--format json" output (SchemaVersion 2) shown on the security page
type trivyReport struct {
	ArtifactType string
	Results      []struct {
		Target          string
		Class           string
		Vulnerabilities []struct {
			VulnerabilityID  string
			PkgName          string
			InstalledVersion string
			FixedVersion     string
			Severity         string
			Title            string
			PrimaryURL       string
		}
		Misconfigurations []struct {
			ID            string
			Title         string
			Message       string
			Severity      string
			PrimaryURL    string
			Status        string
			CauseMetadata struct {
				StartLine int
			}
		}
		Secrets []struct {
			RuleID    string
			Title     string
			Severity  string
			StartLine int
		}
	}
}

func normalizeSeverity(s string) string {
	s = strings.ToUpper(s)
	if slices.Contains(Severities, s) {
		return s
	}
	return "UNKNOWN"
}

func Parse(r io.Reader) (*Report, error) {
	var raw trivyReport
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}
	// image scans report OS package targets like "alpine 3.22" that are not repository paths
	scannedFiles := raw.ArtifactType == "" || raw.ArtifactType == "filesystem" || raw.ArtifactType == "repository"

	report := &Report{SeverityCounts: map[string]int{}, KindCounts: map[string]int{}}
	add := func(f *Finding) {
		f.Severity = normalizeSeverity(f.Severity)
		report.Findings = append(report.Findings, f)
		report.SeverityCounts[f.Severity]++
		report.KindCounts[f.Kind]++
	}
	for _, res := range raw.Results {
		isFile := scannedFiles && res.Class != "os-pkgs"
		for _, v := range res.Vulnerabilities {
			add(&Finding{
				Kind: KindVulnerability, ID: v.VulnerabilityID, Title: cmp.Or(v.Title, v.VulnerabilityID), Severity: v.Severity,
				Target: res.Target, IsFile: isFile, Package: v.PkgName, Installed: v.InstalledVersion, Fixed: v.FixedVersion, URL: v.PrimaryURL,
			})
		}
		for _, m := range res.Misconfigurations {
			if m.Status != "" && m.Status != "FAIL" {
				continue
			}
			add(&Finding{
				Kind: KindMisconfig, ID: m.ID, Title: cmp.Or(m.Title, m.Message), Severity: m.Severity,
				Target: res.Target, IsFile: isFile, Line: m.CauseMetadata.StartLine, URL: m.PrimaryURL,
			})
		}
		for _, s := range res.Secrets { // the matched text is left out, it is the secret itself
			add(&Finding{
				Kind: KindSecret, ID: s.RuleID, Title: cmp.Or(s.Title, s.RuleID), Severity: s.Severity,
				Target: res.Target, IsFile: isFile, Line: s.StartLine,
			})
		}
	}
	slices.SortStableFunc(report.Findings, func(a, b *Finding) int {
		return cmp.Or(
			cmp.Compare(slices.Index(Severities, a.Severity), slices.Index(Severities, b.Severity)),
			cmp.Compare(slices.Index(Kinds, a.Kind), slices.Index(Kinds, b.Kind)),
			strings.Compare(a.Target, b.Target),
			strings.Compare(a.ID, b.ID),
		)
	})
	return report, nil
}

// Filter returns the findings matching kind and severity, an empty value matches all
func (r *Report) Filter(kind, severity string) []*Finding {
	if kind == "" && severity == "" {
		return r.Findings
	}
	return slices.DeleteFunc(slices.Clone(r.Findings), func(f *Finding) bool {
		return (kind != "" && f.Kind != kind) || (severity != "" && f.Severity != severity)
	})
}
