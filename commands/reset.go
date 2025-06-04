package commands

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
)

func Reset(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	var copaings []*user.Copaing
	gokord.DB.Where("guild_id = ?", i.GuildID).Delete(&copaings)
	if err := resp.IsEphemeral().SetMessage("L'XP a été reset.").Send(); err != nil {
		utils.SendAlert("commands/reset.go - Sending success (all)", err.Error())
	}
}

func ResetUser(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()
	v, ok := optMap["user"]
	if !ok {
		if err := resp.SetMessage("Le user n'a pas été renseigné.").Send(); err != nil {
			utils.SendAlert("commands/reset.go - Copaing not set", err.Error())
		}
		return
	}
	m := v.UserValue(s)
	if m.Bot {
		if err := resp.SetMessage("Les bots n'ont pas de niveau :upside_down:").Send(); err != nil {
			utils.SendAlert("commands/reset.go - Copaing not set", err.Error())
		}
		return
	}
	err := user.GetCopaing(m.ID, i.GuildID).Delete()
	if err != nil {
		utils.SendAlert("commands/reset.go - Copaing not deleted", err.Error(), "discord_id", m.ID, "guild_id", i.GuildID)
		err = resp.SetMessage("Erreur : impossible de reset l'utilisateur").Send()
		if err != nil {
			utils.SendAlert("commands/reset.go - Error deleting", err.Error())
		}
	}
	if err = resp.SetMessage("Le user bien été reset.").Send(); err != nil {
		utils.SendAlert("commands/reset.go - Sending success (user)", err.Error())
	}
}
