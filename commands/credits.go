package commands

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
)

func Credits(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	err := resp.AddEmbed(&discordgo.MessageEmbed{

		Type:        discordgo.EmbedTypeRich,
		Title:       "Crédits",
		Description: "Auteur du bot : @anhgelus (https://github.com/anhgelus)\nLangage : Go 1.24\nLicence : AGPLv3",
		Color:       utils.Success,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "anhgelus/gokord",
				Value:  "v0.10.0 - MPL 2.0",
				Inline: true,
			},
			{
				Name:   "bwmarrin/discordgo",
				Value:  "v0.29.0 - BSD-3-Clause",
				Inline: true,
			},
			{
				Name:   "gorm",
				Value:  "v1.30.0 - MIT",
				Inline: true,
			},
		},
	}).Send()
	if err != nil {
		utils.SendAlert("commands/credits.go - Sending credits", err.Error(), "guild_id", i.GuildID)
	}
}
