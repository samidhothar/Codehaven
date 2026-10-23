// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user_test

import (
	"strconv"
	"strings"
	"testing"

	"gitea.dev/models/unittest"
	user_model "gitea.dev/models/user"
	"gitea.dev/modules/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSavedReplies(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	ctx := t.Context()

	replies, err := user_model.GetSavedReplies(ctx, 2)
	require.NoError(t, err)
	assert.Empty(t, replies)

	require.NoError(t, user_model.AddSavedReply(ctx, 2, user_model.SavedReply{Title: " LGTM ", Content: "Looks good to me."}))
	require.NoError(t, user_model.AddSavedReply(ctx, 2, user_model.SavedReply{Title: "Duplicate", Content: "Closing as a duplicate."}))
	assert.ErrorIs(t, user_model.AddSavedReply(ctx, 2, user_model.SavedReply{Title: "lgtm", Content: "again"}), util.ErrAlreadyExist)
	assert.ErrorIs(t, user_model.AddSavedReply(ctx, 2, user_model.SavedReply{Title: "", Content: "x"}), util.ErrInvalidArgument)
	assert.ErrorIs(t, user_model.AddSavedReply(ctx, 2, user_model.SavedReply{Title: "x", Content: strings.Repeat("a", user_model.MaxSavedReplyContentLength+1)}), util.ErrInvalidArgument)

	replies, err = user_model.GetSavedReplies(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, []user_model.SavedReply{{Title: "LGTM", Content: "Looks good to me."}, {Title: "Duplicate", Content: "Closing as a duplicate."}}, replies)

	require.NoError(t, user_model.DeleteSavedReply(ctx, 2, 0))
	assert.ErrorIs(t, user_model.DeleteSavedReply(ctx, 2, 5), util.ErrNotExist)
	replies, err = user_model.GetSavedReplies(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, []user_model.SavedReply{{Title: "Duplicate", Content: "Closing as a duplicate."}}, replies)

	for i := range user_model.MaxSavedReplies - 1 {
		require.NoError(t, user_model.AddSavedReply(ctx, 2, user_model.SavedReply{Title: strconv.Itoa(i), Content: "c"}))
	}
	assert.ErrorIs(t, user_model.AddSavedReply(ctx, 2, user_model.SavedReply{Title: "t", Content: "c"}), util.ErrInvalidArgument)
}
