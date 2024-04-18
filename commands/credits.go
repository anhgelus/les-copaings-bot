package commands

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
)

func Credits(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := utils.ResponseBuilder{C: s, I: i}
	err := resp.Embeds([]*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Crédits",
			Description: "Auteur du bot : @anhgelus (https://github.com/anhgelus)\nLangage : Go 1.22\nLicence : AGPLv3",
			Color:       utils.Success,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "anhgelus/gokord",
					Value:  "v0.3.0 - MPL 2.0",
					Inline: true,
				},
				{
					Name:   "bwmarrin/discordgo",
					Value:  "v0.28.1 - BSD-3-Clause",
					Inline: true,
				},
				{
					Name:   "pelletier/go-toml/v2",
					Value:  "v2.2.1 - MIT",
					Inline: true,
				},
				{
					Name:   "redis/go-redis/v9",
					Value:  "v9.5.1 - BSD-2-Clause",
					Inline: true,
				},
				{
					Name:   "gorm.io/gorm",
					Value:  "v1.25.9 - MIT",
					Inline: true,
				},
				{
					Name:   "gorm.io/driver/postgres",
					Value:  "v1.5.7 - MIT",
					Inline: true,
				},
				{
					Name:   "other",
					Value:  "Et leurs dépendances !",
					Inline: true,
				},
			},
		},
	}).Send()
	if err != nil {
		utils.SendAlert("commands/credits.go - Sending credits", err.Error(), "guild_id", i.GuildID)
	}
}
