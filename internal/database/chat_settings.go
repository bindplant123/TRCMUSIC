/*
 * ● ArcMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 Team Arc
 */

package database

import (
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	state "main/internal/core/models"
)

type RTMPConfig struct {
	URL string `bson:"rtmp_url"`
	Key string `bson:"rtmp_key"`
}

type ChatSettings struct {
	ChatID             int64      `bson:"_id"`
	ChannelPlayID      int64      `bson:"cplay_id"`
	AuthUsers          []int64    `bson:"auth_users"`
	Language           string     `bson:"language"`
	RTMP               RTMPConfig `bson:"rtmp_config"`
	AssistantIndex     int        `bson:"ass_index,omitempty"`
	ThumbnailsDisabled bool       `bson:"no_thumb"`
	PlayModeAdminsOnly bool       `bson:"play_mode"`
	CommandDelete      bool       `bson:"cmd_delete"`
	Autoplay           bool       `bson:"autoplay"`
	PlayedTracks       []state.Track `bson:"played_tracks,omitempty"`
}

func defaultChatSettings(chatID int64) *ChatSettings {
	return &ChatSettings{
		ChatID:    chatID,
		AuthUsers: []int64{},
	}
}

func AutoplayEnabled(chatID int64) (bool, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return settings.Autoplay, nil
}

func SetAutoplay(chatID int64, enabled bool) error {
	return modifyChatSettings(chatID, func(settings *ChatSettings) bool {
		if settings.Autoplay == enabled {
			return false
		}
		settings.Autoplay = enabled
		return true
	})
}

func RecordPlayedTrack(chatID int64, track *state.Track) error {
	if track == nil || track.URL == "" {
		return nil
	}

	return modifyChatSettings(chatID, func(settings *ChatSettings) bool {
		tracks := make([]state.Track, 0, len(settings.PlayedTracks)+1)
		tracks = append(tracks, *track)
		for _, previous := range settings.PlayedTracks {
			if previous.URL == track.URL || (track.ID != "" && previous.ID == track.ID) {
				continue
			}
			tracks = append(tracks, previous)
			if len(tracks) == 50 {
				break
			}
		}
		settings.PlayedTracks = tracks
		return true
	})
}

func PlayedTracks(chatID int64) ([]*state.Track, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return nil, err
	}

	tracks := make([]*state.Track, 0, len(settings.PlayedTracks))
	for i := range settings.PlayedTracks {
		track := settings.PlayedTracks[i]
		tracks = append(tracks, &track)
	}
	return tracks, nil
}

func getChatSettings(chatID int64) (*ChatSettings, error) {
	cacheKey := "chat_settings_" + strconv.FormatInt(chatID, 10)
	if cached, found := dbCache.Get(cacheKey); found {
		if settings, ok := cached.(*ChatSettings); ok {
			return settings, nil
		}
	}

	ctx, cancel := ctx()
	defer cancel()

	var settings ChatSettings
	err := chatSettingsColl.FindOne(ctx, bson.M{"_id": chatID}).
		Decode(&settings)

	if err == mongo.ErrNoDocuments {
		def := defaultChatSettings(chatID)
		dbCache.Set(cacheKey, def)
		return def, nil
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get chat settings for %d: %w",
			chatID,
			err,
		)
	}

	dbCache.Set(cacheKey, &settings)
	return &settings, nil
}

func updateChatSettings(settings *ChatSettings) error {
	cacheKey := "chat_settings_" + strconv.FormatInt(settings.ChatID, 10)

	ctx, cancel := ctx()
	defer cancel()

	_, err := chatSettingsColl.UpdateOne(
		ctx,
		bson.M{"_id": settings.ChatID},
		bson.M{"$set": settings},
		upsertOpt,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update chat settings for %d: %w",
			settings.ChatID,
			err,
		)
	}

	dbCache.Set(cacheKey, settings)
	return nil
}

func modifyChatSettings(chatID int64, fn func(*ChatSettings) bool) error {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return err
	}

	if fn(settings) {
		return updateChatSettings(settings)
	}

	return nil
}
