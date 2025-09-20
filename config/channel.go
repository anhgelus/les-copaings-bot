package config

import (
	"strings"

	"github.com/anhgelus/gokord/cmd"
	discordgo "github.com/nyttikord/gokord"

	// "github.com/nyttikord/gokord/component"
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

func HandleModifyFallbackChannel(s *discordgo.Session, i *discordgo.InteractionCreate, data *interaction.MessageComponentData, resp *cmd.ResponseBuilder) bool {
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

func HandleModifyDisChannel(s *discordgo.Session, i *discordgo.InteractionCreate, data *interaction.MessageComponentData, resp *cmd.ResponseBuilder) bool {
	cfg := GetGuildConfig(i.GuildID)
	cfg.DisabledChannels = strings.Join(data.Values, ";")
	err := cfg.Save()
	if err != nil {
		s.LogError(err, "Unable to save disabled channel")
		return false
	}
	return true
}
