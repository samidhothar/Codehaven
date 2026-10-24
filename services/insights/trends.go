// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package insights

import (
	"context"
	"fmt"
	"slices"
	"time"

	activities_model "gitea.dev/models/activities"
	"gitea.dev/models/db"

	"xorm.io/builder"
)

const secondsPerWeek = 7 * 24 * 60 * 60

// PeriodWeeks are the periods offered by the trends pages, the first one is the default
var PeriodWeeks = []int{12, 4, 26, 52}

func ParseWeeks(weeks int) int {
	if slices.Contains(PeriodWeeks, weeks) {
		return weeks
	}
	return PeriodWeeks[0]
}

type Week struct {
	Start        int64 `json:"start"`
	IssuesOpened int   `json:"issuesOpened"`
	IssuesClosed int   `json:"issuesClosed"`
	PullsOpened  int   `json:"pullsOpened"`
	PullsMerged  int   `json:"pullsMerged"`
	Pushes       int   `json:"pushes"`
	Contributors int   `json:"contributors"`
}

type Trends struct {
	Weeks        []*Week
	IssuesOpened int
	IssuesClosed int
	PullsOpened  int
	PullsMerged  int
	Contributors int
	// MedianMergeSeconds is negative when nothing was merged in the period
	MedianMergeSeconds int64
}

// weekStarts returns the start (Monday 00:00 UTC) of each of the last n weeks, oldest first
func weekStarts(now time.Time, n int) []int64 {
	n = max(0, min(n, 52))
	now = now.UTC()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	current := time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, time.UTC).Unix()
	starts := make([]int64, n)
	for i := range n {
		starts[i] = current - int64(n-1-i)*secondsPerWeek
	}
	return starts
}

func (t *Trends) weekIndex(unix int64) int {
	if len(t.Weeks) == 0 || unix < t.Weeks[0].Start {
		return -1
	}
	idx := int((unix - t.Weeks[0].Start) / secondsPerWeek)
	if idx >= len(t.Weeks) {
		return -1
	}
	return idx
}

func (t *Trends) week(unix int64) *Week {
	if idx := t.weekIndex(unix); idx >= 0 {
		return t.Weeks[idx]
	}
	return nil
}

func median(values []int64) int64 {
	if len(values) == 0 {
		return -1
	}
	slices.Sort(values)
	mid := len(values) / 2
	if len(values)%2 == 1 {
		return values[mid]
	}
	return (values[mid-1] + values[mid]) / 2
}

type issueRow struct {
	CreatedUnix int64
	ClosedUnix  int64
	IsClosed    bool
	IsPull      bool
}

type mergeRow struct {
	MergedUnix  int64
	CreatedUnix int64
}

type actionRow struct {
	ActUserID   int64
	OpType      activities_model.ActionType
	CreatedUnix int64
}

// Collect counts weekly activity in the repositories matched by repoCond (a condition on the repository table),
// a nil repoCond counts every repository
func Collect(ctx context.Context, repoCond builder.Cond, now time.Time, weeks int) (*Trends, error) {
	weeks = ParseWeeks(weeks)
	t := &Trends{}
	for _, start := range weekStarts(now, weeks) {
		t.Weeks = append(t.Weeks, &Week{Start: start})
	}
	since := t.Weeks[0].Start
	inRepos := func(col string) builder.Cond {
		if repoCond == nil {
			return builder.NewCond()
		}
		return builder.In(col, builder.Select("id").From("repository").Where(repoCond))
	}

	var issues []issueRow
	if err := db.GetEngine(ctx).Table("issue").Cols("created_unix", "closed_unix", "is_closed", "is_pull").
		Where(inRepos("repo_id")).And(builder.Or(builder.Gte{"created_unix": since}, builder.Gte{"closed_unix": since})).
		Find(&issues); err != nil {
		return nil, fmt.Errorf("Collect: issues: %w", err)
	}
	for _, issue := range issues {
		if w := t.week(issue.CreatedUnix); w != nil {
			if issue.IsPull {
				w.PullsOpened++
				t.PullsOpened++
			} else {
				w.IssuesOpened++
				t.IssuesOpened++
			}
		}
		if w := t.week(issue.ClosedUnix); w != nil && issue.IsClosed && !issue.IsPull {
			w.IssuesClosed++
			t.IssuesClosed++
		}
	}

	var merges []mergeRow
	if err := db.GetEngine(ctx).Table("pull_request").Join("INNER", "issue", "issue.id = pull_request.issue_id").
		Select("pull_request.merged_unix, issue.created_unix").
		Where(builder.Eq{"pull_request.has_merged": true}).And(builder.Gte{"pull_request.merged_unix": since}).And(inRepos("issue.repo_id")).
		Find(&merges); err != nil {
		return nil, fmt.Errorf("Collect: merges: %w", err)
	}
	durations := make([]int64, 0, len(merges))
	for _, merge := range merges {
		if w := t.week(merge.MergedUnix); w != nil {
			w.PullsMerged++
			t.PullsMerged++
			durations = append(durations, merge.MergedUnix-merge.CreatedUnix)
		}
	}
	t.MedianMergeSeconds = median(durations)

	var actions []actionRow
	// every action is copied into the feed of each watcher, the actor's own copy counts it once
	if err := db.GetEngine(ctx).Table("action").Cols("act_user_id", "op_type", "created_unix").
		Where(builder.Expr("user_id = act_user_id")).And(builder.Gte{"created_unix": since}).And(inRepos("repo_id")).
		Find(&actions); err != nil {
		return nil, fmt.Errorf("Collect: actions: %w", err)
	}
	weekUsers := make([]map[int64]struct{}, len(t.Weeks))
	allUsers := map[int64]struct{}{}
	for _, action := range actions {
		idx := t.weekIndex(action.CreatedUnix)
		if idx < 0 {
			continue
		}
		if action.OpType == activities_model.ActionCommitRepo {
			t.Weeks[idx].Pushes++
		}
		if weekUsers[idx] == nil {
			weekUsers[idx] = map[int64]struct{}{}
		}
		weekUsers[idx][action.ActUserID] = struct{}{}
		allUsers[action.ActUserID] = struct{}{}
	}
	for i, users := range weekUsers {
		t.Weeks[i].Contributors = len(users)
	}
	t.Contributors = len(allUsers)
	return t, nil
}
