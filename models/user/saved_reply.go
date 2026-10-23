// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"context"
	"strings"
	"unicode/utf8"

	"gitea.dev/modules/util"
)

// Limits keep the JSON setting value within MySQL's 64 KiB TEXT column
const (
	MaxSavedReplies             = 20
	MaxSavedReplyTitleLength    = 100
	MaxSavedReplyContentLength  = 2000
	SettingsKeyEditorSavedReply = "editor.saved_replies"
)

type SavedReply struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func GetSavedReplies(ctx context.Context, userID int64) ([]SavedReply, error) {
	return GetUserSettingJSON(ctx, userID, SettingsKeyEditorSavedReply, []SavedReply{})
}

func AddSavedReply(ctx context.Context, userID int64, reply SavedReply) error {
	reply.Title = strings.TrimSpace(reply.Title)
	if reply.Title == "" || strings.TrimSpace(reply.Content) == "" {
		return util.NewInvalidArgumentErrorf("saved reply title and content are required")
	}
	if utf8.RuneCountInString(reply.Title) > MaxSavedReplyTitleLength || utf8.RuneCountInString(reply.Content) > MaxSavedReplyContentLength {
		return util.NewInvalidArgumentErrorf("saved reply is too long")
	}
	replies, err := GetSavedReplies(ctx, userID)
	if err != nil {
		return err
	}
	if len(replies) >= MaxSavedReplies {
		return util.NewInvalidArgumentErrorf("saved reply limit reached")
	}
	// titles pick the reply in the slash command menu, so they must be unique; this also absorbs double-submits
	for _, existing := range replies {
		if strings.EqualFold(existing.Title, reply.Title) {
			return util.NewAlreadyExistErrorf("saved reply %q already exists", reply.Title)
		}
	}
	return SetUserSettingJSON(ctx, userID, SettingsKeyEditorSavedReply, append(replies, reply))
}

func DeleteSavedReply(ctx context.Context, userID int64, index int) error {
	replies, err := GetSavedReplies(ctx, userID)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(replies) {
		return util.NewNotExistErrorf("saved reply does not exist")
	}
	return SetUserSettingJSON(ctx, userID, SettingsKeyEditorSavedReply, append(replies[:index], replies[index+1:]...))
}
