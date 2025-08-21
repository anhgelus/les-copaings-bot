package commands

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
)

func Credits(_ *discordgo.Session, i *discordgo.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	msg := "**Les Copaings**, le bot gérant les serveurs privés de [anhgelus](<https://anhgelus.world/>).\n"
	msg += "Code source : <https://git.anhgelus.world/anhgelus/les-copaings-bot>\n\n"
	msg += "Host du bot : " + gokord.BaseCfg.GetAuthor() + "\n\n"
	msg += "Utilise :\n- [anhgelus/gokord](<https://github.com/anhgelus/gokord>)"
	err := resp.SetMessage(msg).Send()
	if err != nil {
		logger.Alert("commands/credits.go - Sending credits", err.Error(), "guild_id", i.GuildID)
	}
}
