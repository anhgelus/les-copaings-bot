package commands

import (
	"context"
	"fmt"
	"sync"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/interaction"
)

func Top(ctx context.Context, dg bot.Session, i *interaction.ApplicationCommand) {
	embeds := make([]*channel.MessageEmbed, 3)
	var wg sync.WaitGroup

	fn := func(str string, n uint, d int, id int) {
		tops := user.GetBestXP(ctx, i.GuildID, n, d)
		embeds[id] = &channel.MessageEmbed{
			Title:       str,
			Description: genTopsMessage(tops),
			Color:       0x10E6AD,
		}
	}

	cfg := config.GetGuildConfig(ctx, i.GuildID)
	if cfg.DaysXPRemains > 30 {
		wg.Go(func() {
			fn(fmt.Sprintf("Top %d jours", cfg.DaysXPRemains), 10, -1, 0)
		})
	}

	wg.Go(func() {
		fn("Top 30 jours", 5, 30, 1)
	})
	wg.Go(func() {
		fn("Top 7 jours", 5, 7, 2)
	})

	wg.Wait()
	resp := interaction.NewMessageResponse()
	if cfg.DaysXPRemains > 30 {
		resp.AddEmbed(embeds[0]).
			AddEmbed(embeds[1]).
			AddEmbed(embeds[2])
	} else {
		resp.AddEmbed(embeds[1]).
			AddEmbed(embeds[2])
	}
	err := dg.InteractionAPI().Respond(i.Interaction, resp.Response()).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("sending response top", "error", err)
	}
}

func genTopsMessage(tops []user.CopaingCached) string {
	msg := ""
	for i, c := range tops {
		msg += fmt.Sprintf("%d. **<@%s>** - niveau %d", i+1, c.DiscordID, exp.Level(c.XP))
		if i != len(tops)-1 {
			msg += "\n"
		}
	}
	return msg
}
