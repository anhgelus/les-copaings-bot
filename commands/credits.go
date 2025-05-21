package commands

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
)

func Credits(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := utils.NewResponseBuilder(s, i)
	err := resp.Embeds([]*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Crédits",
			Description: "Auteur du bot : @anhgelus (https://github.com/anhgelus)\nLangage : Go 1.24\nLicence : AGPLv3",
			Color:       utils.Success,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "anhgelus/gokord",
					Value:  "v0.6.3 - MPL 2.0",
					Inline: true,
				},
				{
					Name:   "bwmarrin/discordgo",
					Value:  "v0.28.1 - BSD-3-Clause",
					Inline: true,
				},
				{
					Name:   "redis/go-redis/v9",
					Value:  "v9.8.0 - BSD-2-Clause",
					Inline: true,
				},
			},
		},
	}).Send()
	if err != nil {
		utils.SendAlert("commands/credits.go - Sending credits", err.Error(), "guild_id", i.GuildID)
	}
}
