package commands

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
)

func Reset(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var copaings []*user.Copaing
	gokord.DB.Where("guild_id = ?", i.GuildID).Delete(&copaings)
	resp := utils.ResponseBuilder{C: s, I: i}
	if err := resp.IsEphemeral().Message("L'XP a été reset.").Send(); err != nil {
		utils.SendAlert("commands/reset.go - Sending success (all)", err.Error())
	}
}

func ResetUser(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := utils.ResponseBuilder{C: s, I: i}
	resp.IsEphemeral()
	optMap := utils.GenerateOptionMap(i)
	v, ok := optMap["user"]
	if !ok {
		if err := resp.Message("Le user n'a pas été renseigné.").Send(); err != nil {
			utils.SendAlert("commands/reset.go - Copaing not set", err.Error())
		}
		return
	}
	m := v.UserValue(s)
	if m.Bot {
		if err := resp.Message("Les bots n'ont pas de niveau :upside_down:").Send(); err != nil {
			utils.SendAlert("commands/reset.go - Copaing not set", err.Error())
		}
		return
	}
	user.GetCopaing(m.ID, i.GuildID).Reset()
	if err := resp.Message("Le user bien été reset.").Send(); err != nil {
		utils.SendAlert("commands/reset.go - Sending success (user)", err.Error())
	}
}
