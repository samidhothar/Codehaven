// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package insights

import (
	"testing"
	"time"

	"gitea.dev/models/db"
	"gitea.dev/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestWeekStarts(t *testing.T) {
	saturday := time.Date(2026, 10, 24, 15, 30, 0, 0, time.UTC)
	assert.Equal(t, []int64{
		time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC).Unix(),
		time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC).Unix(),
		time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC).Unix(),
	}, weekStarts(saturday, 3))
	monday := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, monday.Unix(), weekStarts(monday, 1)[0])
	assert.Len(t, weekStarts(monday, 1<<30), 52)
	assert.Empty(t, weekStarts(monday, -1))
}

func TestMedian(t *testing.T) {
	assert.EqualValues(t, -1, median(nil))
	assert.EqualValues(t, 5, median([]int64{9, 1, 5}))
	assert.EqualValues(t, 4, median([]int64{2, 6, 1, 9}))
}

func TestParseWeeks(t *testing.T) {
	assert.Equal(t, 26, ParseWeeks(26))
	assert.Equal(t, 12, ParseWeeks(0))
	assert.Equal(t, 12, ParseWeeks(1000))
}

func TestCollect(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	now := time.Date(2000, 1, 20, 0, 0, 0, 0, time.UTC) // the fixtures are created around 2000-01-01
	since := weekStarts(now, 4)[0]

	all, err := Collect(t.Context(), nil, now, 4)
	require.NoError(t, err)
	require.Len(t, all.Weeks, 4)
	sum := 0
	for _, w := range all.Weeks {
		sum += w.IssuesOpened
	}
	assert.Equal(t, all.IssuesOpened, sum)

	repo1, err := Collect(t.Context(), builder.Eq{"`repository`.id": 1}, now, 4)
	require.NoError(t, err)
	expected, err := db.GetEngine(t.Context()).Table("issue").Where(builder.Eq{"repo_id": 1, "is_pull": false}).And(builder.Gte{"created_unix": since}).Count()
	require.NoError(t, err)
	assert.Positive(t, expected)
	assert.EqualValues(t, expected, repo1.IssuesOpened)
	assert.LessOrEqual(t, repo1.IssuesOpened, all.IssuesOpened)
}
