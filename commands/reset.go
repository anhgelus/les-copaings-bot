package commands

import (
	"context"

	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/interaction"
)

func Reset(ctx context.Context, dg bot.Session, i *interaction.ApplicationCommand) {
	var copaings []*user.Copaing
	//TODO: delete everything from cache
	gokord.DB.Where("guild_id = ?", i.GuildID).Delete(&copaings)
	resp := interaction.NewMessageResponse().IsEphemeral().Message("L'XP a été reset.").Response()
	err := dg.InteractionAPI().Respond(i.Interaction, resp).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("sending reset success", "error", err)
	}
}

func ResetUser(ctx context.Context, dg bot.Session, i *interaction.ApplicationCommand) {
	resp := interaction.NewMessageResponse().IsEphemeral()
	defer func() {
		err := dg.InteractionAPI().Respond(i.Interaction, resp.Response()).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("replying to interaction", "error", err)
		}
	}()
	opts := i.OptionMap()
	v, ok := opts["user"]
	if !ok {
		resp.Message("Le user n'a pas été renseigné.")
		return
	}
	m := v.UserValue(ctx)
	if m.Bot {
		resp.Message("Les bots n'ont pas de niveau :upside_down:")
		return
	}
	err := user.GetCopaing(ctx, m.ID, i.GuildID).Delete(ctx)
	if err != nil {
		bot.Logger(ctx).Error("deleting copaing", "error", err, "user", m.Username, "guild", i.GuildID)
		resp.Message("Erreur : impossible de reset l'utilisateur")
		return
	}
	resp.Message("Le user bien été reset.")
}
