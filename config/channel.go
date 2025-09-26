package config

import (
	"strings"

	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/event"

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

func HandleModifyFallbackChannel(s bot.Session, i *event.InteractionCreate, data *interaction.MessageComponentData, _ *cmd.ResponseBuilder) bool {
	cfg := GetGuildConfig(i.GuildID)
	var channelID string
	if len(data.Values) > 0 {
		channelID = data.Values[0]
	}
	cfg.FallbackChannel = channelID
	err := cfg.Save()
	if err != nil {
		s.LogError(err, "Saving fallback channel")
		return false
	}
	return true
}

func HandleModifyDisChannel(s bot.Session, i *event.InteractionCreate, data *interaction.MessageComponentData, _ *cmd.ResponseBuilder) bool {
	cfg := GetGuildConfig(i.GuildID)
	cfg.DisabledChannels = strings.Join(data.Values, ";")
	err := cfg.Save()
	if err != nil {
		s.LogError(err, "Unable to save disabled channel")
		return false
	}
	return true
}
