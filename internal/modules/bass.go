/*
 * ● ArcMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 Team Arc
 */

package modules

import (
	"main/internal/core"
	"github.com/amarnathcjd/gogram/telegram"
)

func bassHandler(ctx *telegram.NewMessage) error {
	chatID := ctx.Chat.ID
	room, exists := core.GetRoom(chatID, nil, false)

	if !exists || room == nil || !room.IsPlaying() {
		return ctx.Reply(F(chatID, "NOT_PLAYING"))
	}

	isBass := room.ToggleBass()
	if isBass {
		return ctx.Reply(F(chatID, "BASS_ENABLED"))
	}
	return ctx.Reply(F(chatID, "BASS_DISABLED"))
}
