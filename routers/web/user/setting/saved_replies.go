// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"errors"
	"net/http"

	user_model "gitea.dev/models/user"
	"gitea.dev/modules/setting"
	"gitea.dev/modules/templates"
	"gitea.dev/modules/util"
	"gitea.dev/services/context"
)

const tplSettingsSavedReplies templates.TplName = "user/settings/saved_replies"

func SavedReplies(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.saved_replies")
	ctx.Data["PageIsSettingsSavedReplies"] = true

	replies, err := user_model.GetSavedReplies(ctx, ctx.Doer.ID)
	if err != nil {
		ctx.ServerError("GetSavedReplies", err)
		return
	}
	ctx.Data["SavedReplies"] = replies
	ctx.Data["CanAddSavedReply"] = len(replies) < user_model.MaxSavedReplies
	ctx.Data["MaxSavedReplyTitleLength"] = user_model.MaxSavedReplyTitleLength
	ctx.Data["MaxSavedReplyContentLength"] = user_model.MaxSavedReplyContentLength
	ctx.HTML(http.StatusOK, tplSettingsSavedReplies)
}

func SavedRepliesPost(ctx *context.Context) {
	var err error
	switch ctx.FormString("action") {
	case "add":
		err = user_model.AddSavedReply(ctx, ctx.Doer.ID, user_model.SavedReply{Title: ctx.FormString("title"), Content: ctx.FormString("content")})
	case "delete":
		err = user_model.DeleteSavedReply(ctx, ctx.Doer.ID, ctx.FormInt("index"))
	default:
		ctx.HTTPError(http.StatusBadRequest)
		return
	}

	switch {
	case err == nil:
		ctx.JSONRedirect(setting.AppSubURL + "/user/settings/replies")
	case errors.Is(err, util.ErrAlreadyExist):
		ctx.JSONError(ctx.Tr("settings.saved_replies.exists"))
	case errors.Is(err, util.ErrInvalidArgument):
		ctx.JSONError(ctx.Tr("settings.saved_replies.invalid", user_model.MaxSavedReplyTitleLength, user_model.MaxSavedReplyContentLength, user_model.MaxSavedReplies))
	case errors.Is(err, util.ErrNotExist):
		ctx.JSONError(ctx.Tr("error.not_found"))
	default:
		ctx.ServerError("SavedRepliesPost", err)
	}
}

// SavedRepliesList serves the editor's slash command menu
func SavedRepliesList(ctx *context.Context) {
	replies, err := user_model.GetSavedReplies(ctx, ctx.Doer.ID)
	if err != nil {
		ctx.ServerError("GetSavedReplies", err)
		return
	}
	ctx.JSON(http.StatusOK, replies)
}
