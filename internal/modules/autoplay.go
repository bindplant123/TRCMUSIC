package modules

import (
	"context"
	"fmt"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
)

func init() {
	helpTexts["/autoplay"] = `<i>Automatically continue music in this group.</i>

<b>/autoplay</b> — Toggle autoplay on or off
<b>/autoplay off</b> — Turn autoplay off

When enabled, the bot replays this group's saved song history. If there is no history yet, it starts with ten popular Hindi songs.`
}

func autoplayHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()
	args := strings.Fields(m.Text())
	enabled, err := database.AutoplayEnabled(chatID)
	if err != nil {
		m.Reply("Unable to read autoplay settings: " + utils.EscapeHTML(err.Error()))
		return tg.ErrEndGroup
	}

	wantsOff := len(args) > 1 && strings.EqualFold(args[1], "off")
	if wantsOff || (len(args) == 1 && enabled) {
		if err := database.SetAutoplay(chatID, false); err != nil {
			m.Reply("Unable to update autoplay: " + utils.EscapeHTML(err.Error()))
			return tg.ErrEndGroup
		}
		m.Reply("<b>Autoplay disabled.</b>", &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	if err := database.SetAutoplay(chatID, true); err != nil {
		m.Reply("Unable to update autoplay: " + utils.EscapeHTML(err.Error()))
		return tg.ErrEndGroup
	}

	tracks, err := autoplayTracks(chatID)
	if err != nil {
		m.Reply("Unable to load autoplay tracks: " + utils.EscapeHTML(err.Error()))
		return tg.ErrEndGroup
	}

	r, err := getEffectiveRoom(m, false)
	if err != nil {
		return tg.ErrEndGroup
	}
	r.Parse()
	mention := utils.MentionHTML(m.Sender)
	for _, track := range tracks {
		track.Requester = mention
	}

	if r.IsActiveChat() {
		available := config.QueueLimit - len(r.Queue())
		if available < 1 {
			m.Reply("<b>Autoplay enabled.</b> The queue is currently full.", &tg.SendOptions{ParseMode: "HTML"})
			return tg.ErrEndGroup
		}
		if len(tracks) > available {
			tracks = tracks[:available]
		}
		r.AddTracksToQueue(tracks)
		m.Reply(fmt.Sprintf("<b>Autoplay enabled.</b> Added %d tracks to this group's queue.", len(tracks)), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	chatState, err := core.GetChatState(chatID)
	if err != nil {
		m.Reply(getErrorMessage(chatID, err))
		return tg.ErrEndGroup
	}
	activeVC, err := chatState.IsActiveVC(false)
	if err != nil || !activeVC {
		m.Reply(F(chatID, "err_no_active_voicechat"))
		return tg.ErrEndGroup
	}
	if err := chatState.TryJoin(); err != nil {
		m.Reply(getErrorMessage(chatID, err))
		return tg.ErrEndGroup
	}

	statusMsg, _ := m.Reply(F(chatID, "stream_downloading_next"))
	filePath, err := platforms.ResolvePlaybackInput(context.Background(), tracks[0], statusMsg)
	if err != nil {
		utils.EOR(statusMsg, "Autoplay could not start: "+utils.EscapeHTML(err.Error()))
		return tg.ErrEndGroup
	}
	if err := playTrackWithRetry(r, tracks[0], filePath, false, statusMsg); err != nil {
		return err
	}
	database.RecordPlayedTrack(chatID, tracks[0])
	if len(tracks) > 1 {
		r.AddTracksToQueue(tracks[1:])
	}

	opt := &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, false),
	}
	utils.EOR(statusMsg, F(chatID, "stream_now_playing", locales.Arg{
		"url":      tracks[0].URL,
		"title":    utils.EscapeHTML(utils.ShortTitle(tracks[0].Title, 25)),
		"duration": formatDuration(tracks[0].Duration),
		"by":       mention,
	}), opt)
	r.SetStatusMsg(statusMsg)
	return tg.ErrEndGroup
}

func autoplayTracks(chatID int64) ([]*state.Track, error) {
	tracks, err := database.PlayedTracks(chatID)
	if err != nil {
		return nil, err
	}
	if len(tracks) > 0 {
		return tracks, nil
	}
	return platforms.GetYouTubeTracks("top 10 best Hindi songs", 10)
}
