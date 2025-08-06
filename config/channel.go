package config

import (
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/component"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
)

const (
	ModifyDisChannel      = "disabled_channel"
	ModifyFallbackChannel = "fallback_channel"
	FallbackChannelSet    = "fallback_channel_set"
)

func HandleModifyFallbackChannel(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	err := resp.SetMessage("Salon de repli...").SetComponents(component.New().Add(component.NewActionRow().Add(
		component.NewChannelSelect(FallbackChannelSet).AddChannelType(discordgo.ChannelTypeGuildText),
	))).Send()
	if err != nil {
		logger.Alert("config/channel.go - Sending channel list", err.Error())
	}
}

func HandleFallbackChannelSet(_ *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()

	cfg := GetGuildConfig(i.GuildID)
	channelID := data.Values[0]

	cfg.FallbackChannel = channelID
	err := cfg.Save()
	if err != nil {
		logger.Alert("config/channel.go - Saving fallback channel", err.Error())
		if err = resp.SetMessage("Erreur lors de la sauvegarde du salon").Send(); err != nil {
			logger.Alert("config/channel.go - Sending error while saving channel", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Salon sauvegardé.").Send(); err != nil {
		logger.Alert("config/channel.go - Sending channel saved", err.Error())
	}
}
