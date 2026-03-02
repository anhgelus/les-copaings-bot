package commands

import (
	"context"
	"fmt"

	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/guild"
	"github.com/nyttikord/gokord/interaction"
)

func Rank(ctx context.Context, dg bot.Session, i *interaction.ApplicationCommand) {
	c := user.GetCopaing(ctx, i.Member.User.ID, i.GuildID) // current user = member who used /rank
	msg := "Votre niveau"
	m := i.Member
	opts := i.OptionMap()
	resp := interaction.NewMessageResponse()
	defer func() {
		err := interaction.Respond(i.Interaction, resp.Response()).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("replying to interaction", "error", err)
		}
	}()
	var err error
	if v, ok := opts["copaing"]; ok {
		u := v.UserValue(ctx)
		if u.Bot {
			resp.IsEphemeral().Message("Imagine si les bots avaient un niveau :rolling_eyes:")
			return
		}
		m, err = guild.GetMember(i.GuildID, u.ID).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("fetching guild member", "error", err, "user", u.Username, "guild", i.GuildID)
			resp.IsEphemeral().Message("Erreur : impossible de récupérer le membre")
			return
		}
		c = user.GetCopaing(ctx, u.ID, i.GuildID) // current user = member targeted by member who wrote /rank
		msg = fmt.Sprintf("Le niveau de %s", m.DisplayName())
	}
	xp := c.XP
	lvl := exp.Level(xp)
	nxtLvlXP := exp.LevelXP(lvl + 1)
	resp.Message(fmt.Sprintf(
		"%s : **%d**\n> XP : %d\n> Prochain niveau dans %d XP",
		msg, lvl, xp, nxtLvlXP-xp,
	))
}
