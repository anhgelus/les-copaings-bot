package commands

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
)

func Reset(_ *discordgo.Session, i *discordgo.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	var copaings []*user.Copaing
	gokord.DB.Where("guild_id = ?", i.GuildID).Delete(&copaings)
	if err := resp.IsEphemeral().SetMessage("L'XP a été reset.").Send(); err != nil {
		logger.Alert("commands/reset.go - Sending success (all)", err.Error())
	}
}

func ResetUser(s *discordgo.Session, i *discordgo.InteractionCreate, optMap cmd.OptionMap, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	v, ok := optMap["user"]
	if !ok {
		if err := resp.SetMessage("Le user n'a pas été renseigné.").Send(); err != nil {
			logger.Alert("commands/reset.go - Copaing not set", err.Error())
		}
		return
	}
	m := v.UserValue(s)
	if m.Bot {
		if err := resp.SetMessage("Les bots n'ont pas de niveau :upside_down:").Send(); err != nil {
			logger.Alert("commands/reset.go - Copaing not set", err.Error())
		}
		return
	}
	err := user.GetCopaing(m.ID, i.GuildID).Delete()
	if err != nil {
		logger.Alert("commands/reset.go - Copaing not deleted", err.Error(), "discord_id", m.ID, "guild_id", i.GuildID)
		err = resp.SetMessage("Erreur : impossible de reset l'utilisateur").Send()
		if err != nil {
			logger.Alert("commands/reset.go - Error deleting", err.Error())
		}
	}
	if err = resp.SetMessage("Le user bien été reset.").Send(); err != nil {
		logger.Alert("commands/reset.go - Sending success (user)", err.Error())
	}
}
