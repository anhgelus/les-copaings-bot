package commands

import (
	"context"

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
		s.Logger().Error("sending reset success", "error", err)
	}
}

func ResetUser(ctx context.Context) func(s bot.Session, i *event.InteractionCreate, optMap cmd.OptionMap, resp *cmd.ResponseBuilder) {
	return func(s bot.Session, i *event.InteractionCreate, optMap cmd.OptionMap, resp *cmd.ResponseBuilder) {
		resp.IsEphemeral()
		v, ok := optMap["user"]
		if !ok {
			if err := resp.SetMessage("Le user n'a pas été renseigné.").Send(); err != nil {
				s.Logger().Error("sending error copaing not set", "error", err)
			}
			return
		}
		m := v.UserValue(s.UserAPI())
		if m.Bot {
			if err := resp.SetMessage("Les bots n'ont pas de niveau :upside_down:").Send(); err != nil {
				s.Logger().Error("sending error bot does not have xp", "error", err)
			}
			return
		}
		err := user.GetCopaing(ctx, m.ID, i.GuildID).Copaing(ctx).Delete(ctx)
		if err != nil {
			s.Logger().Error("deleting copaing", "error", err, "user", m.Username, "guild", i.GuildID)
			err = resp.SetMessage("Erreur : impossible de reset l'utilisateur").Send()
			if err != nil {
				s.Logger().Error("sending error while deleting", "error", err)
			}
		}
		if err = resp.SetMessage("Le user bien été reset.").Send(); err != nil {
			s.Logger().Error("sending reset success", "error", err)
		}
	}
}
