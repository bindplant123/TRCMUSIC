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

func init() {
	core.AddCommandHandler("bass", core.Command{
		Handler: bassHandler,
		Admin:   true,
	})
}

func bassHandler(ctx *telegram.NewMessage) error {
	chatID := ctx.Chat.ID
	room := core.GetRoom(chatID)

	if room == nil || !room.IsPlaying() {
		return F(ctx, "NOT_PLAYING")
	}

	isBass := room.ToggleBass()
	if isBass {
		return F(ctx, "BASS_ENABLED")
	}
	return F(ctx, "BASS_DISABLED")
}
