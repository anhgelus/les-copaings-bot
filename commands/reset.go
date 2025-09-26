package commands

import (
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/event"
)

func Reset(s bot.Session, i *event.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	var copaings []*user.Copaing
	gokord.DB.Where("guild_id = ?", i.GuildID).Delete(&copaings)
	if err := resp.IsEphemeral().SetMessage("L'XP a été reset.").Send(); err != nil {
		s.LogError(err, "sending reset success")
	}
}

func ResetUser(s bot.Session, i *event.InteractionCreate, optMap cmd.OptionMap, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	v, ok := optMap["user"]
	if !ok {
		if err := resp.SetMessage("Le user n'a pas été renseigné.").Send(); err != nil {
			s.LogError(err, "sending error copaing not set")
		}
		return
	}
	m := v.UserValue(s.UserAPI())
	if m.Bot {
		if err := resp.SetMessage("Les bots n'ont pas de niveau :upside_down:").Send(); err != nil {
			s.LogError(err, "sending error bot does not have xp")
		}
		return
	}
	err := user.GetCopaing(m.ID, i.GuildID).Delete()
	if err != nil {
		s.LogError(err, "deleting copaings %s in %s", m.Username, i.GuildID)
		err = resp.SetMessage("Erreur : impossible de reset l'utilisateur").Send()
		if err != nil {
			s.LogError(err, "sending error while deleting")
		}
	}
	if err = resp.SetMessage("Le user bien été reset.").Send(); err != nil {
		s.LogError(err, "sending reset success")
	}
}
