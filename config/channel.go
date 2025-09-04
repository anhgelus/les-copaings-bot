package config

import (
	"strings"

	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/component"
	"github.com/anhgelus/gokord/logger"
	discordgo "github.com/nyttikord/gokord"
)

const (
	ModifyFallbackChannel = "fallback_channel"
	FallbackChannelSet    = "fallback_channel_set"

	ModifyDisChannel = "disabled_channel"
	DisChannelAdd    = "disabled_channel_add"
	DisChannelAddSet = "disabled_channel_add_set"
	DisChannelDel    = "disabled_channel_del"
	DisChannelDelSet = "disabled_channel_del_set"
)

func HandleModifyFallbackChannel(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	err := resp.IsEphemeral().SetComponents(component.New().Add(component.NewActionRow().Add(
		component.NewChannelSelect(FallbackChannelSet).AddChannelType(discordgo.ChannelTypeGuildText),
	))).Send()
	if err != nil {
		logger.Alert("config/channel.go - Sending channel list for fallback", err.Error())
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

func HandleModifyDisChannel(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	err := resp.IsEphemeral().SetComponents(component.New().Add(component.NewActionRow().
		Add(
			component.NewButton(DisChannelAdd, discordgo.PrimaryButton).
				SetLabel("Désactiver un salon").
				SetEmoji(&discordgo.ComponentEmoji{Name: "⬇️"}),
		).
		Add(
			component.NewButton(DisChannelDel, discordgo.DangerButton).
				SetLabel("Réactiver un salon").
				SetEmoji(&discordgo.ComponentEmoji{Name: "⬆️"}),
		),
	)).Send()
	if err != nil {
		logger.Alert("config/channel.go - Sending action type", err.Error())
	}
}

func HandleDisChannel(_ *discordgo.Session, _ *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral().SetMessage("Salon à désactiver...")
	cID := DisChannelAddSet
	if data.CustomID == DisChannelDel {
		resp.SetMessage("Salon à réactiver...")
		cID = DisChannelDelSet
	}
	err := resp.SetComponents(component.New().Add(component.NewActionRow().Add(component.NewChannelSelect(cID)))).Send()
	if err != nil {
		logger.Alert("config/channel.go - Sending channel list for disable", err.Error())
	}
}

func HandleDisChannelAddSet(_ *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	cfg := GetGuildConfig(i.GuildID)
	id := data.Values[0]
	if strings.Contains(cfg.DisabledChannels, id) {
		err := resp.SetMessage("Le salon est déjà dans la liste des salons désactivés").Send()
		if err != nil {
			logger.Alert("commands/config.go - Channel already disabled", err.Error())
		}
		return
	}
	cfg.DisabledChannels += id + ";"
	if err := cfg.Save(); err != nil {
		logger.Alert("commands/config.go - Saving config disable add", err.Error())
		if err = resp.SetMessage("Il y a eu une erreur lors de la modification de de la base de données.").Send(); err != nil {
			logger.Alert("config/channel.go - Sending error while saving config", err.Error())
		}
	}
	if err := resp.SetMessage("Modification sauvegardé.").Send(); err != nil {
		logger.Alert("commands/config.go - Modification saved message disable add", err.Error())
	}
}

func HandleDisChannelDelSet(_ *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	cfg := GetGuildConfig(i.GuildID)
	id := data.Values[0]
	if !strings.Contains(cfg.DisabledChannels, id) {
		err := resp.SetMessage("Le salon n'est pas désactivé").Send()
		if err != nil {
			logger.Alert("commands/config.go - Channel not disabled", err.Error())
		}
		return
	}
	cfg.DisabledChannels = strings.ReplaceAll(cfg.DisabledChannels, id+";", "")
	if err := cfg.Save(); err != nil {
		logger.Alert("commands/config.go - Saving config disable del", err.Error())
		if err = resp.SetMessage("Il y a eu une erreur lors de la modification de de la base de données.").Send(); err != nil {
			logger.Alert("config/channel.go - Sending error while saving config", err.Error())
		}
	}
	if err := resp.SetMessage("Modification sauvegardé.").Send(); err != nil {
		logger.Alert("commands/config.go - Modification saved message disable del", err.Error())
	}
}
