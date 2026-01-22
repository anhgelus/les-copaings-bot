package config

import (
	"context"
	"strings"

	"github.com/nyttikord/gokord/bot"

	"github.com/nyttikord/gokord/interaction"
)

const (
	ModifyFallbackChannel = "fallback_channel"

	ModifyDisChannel = "disabled_channel"
	DisChannelAdd    = "disabled_channel_add"
	DisChannelAddSet = "disabled_channel_add_set"
	DisChannelDel    = "disabled_channel_del"
	DisChannelDelSet = "disabled_channel_del_set"
)

func HandleModifyFallbackChannel(ctx context.Context, dg bot.Session, i *interaction.MessageComponent) bool {
	cfg := GetGuildConfig(i.GuildID)
	var channelID string
	if len(i.Data.Values) > 0 {
		channelID = i.Data.Values[0]
	}
	cfg.FallbackChannel = channelID
	err := cfg.Save()
	if err != nil {
		bot.Logger(ctx).Error("saving fallback channel", "error", err)
		return false
	}
	return true
}

func HandleModifyDisChannel(ctx context.Context, dg bot.Session, i *interaction.MessageComponent) bool {
	cfg := GetGuildConfig(i.GuildID)
	cfg.DisabledChannels = strings.Join(i.Data.Values, ";")
	err := cfg.Save()
	if err != nil {
		bot.Logger(ctx).Error("unable to save disabled channel", "error", err)
		return false
	}
	return true
}
