/*
 * ● ArcMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 Team Arc
 */

package modules

import (
	"github.com/amarnathcjd/gogram/telegram"
)

func bassHandler(m *telegram.NewMessage) error {
	r, err := getEffectiveRoom(m, false)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()

	if !r.IsActiveChat() || r.Track() == nil {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.ErrEndGroup
	}

	isBass := r.ToggleBass()
	if isBass {
		m.Reply(F(chatID, "BASS_ENABLED"))
	} else {
		m.Reply(F(chatID, "BASS_DISABLED"))
	}
	return telegram.ErrEndGroup
}
