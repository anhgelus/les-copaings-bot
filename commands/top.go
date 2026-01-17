package commands

import (
	"fmt"
	"sync"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/event"
)

func Top(s bot.Session, i *event.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	err := resp.IsDeferred().Send()
	if err != nil {
		s.Logger().Error("sending deferred", "error", err)
		return
	}
	embeds := make([]*channel.MessageEmbed, 3)
	wg := sync.WaitGroup{}

	fn := func(str string, n uint, d int, id int) {
		defer wg.Done()
		tops, err := user.GetBestXP(s.Logger(), i.GuildID, n, d)
		if err != nil {
			s.Logger().Error("fetching best xp", "error", err, "n", n, "d", d, "id", id, "guild", i.GuildID)
			embeds[id] = &channel.MessageEmbed{
				Title:       str,
				Description: "Erreur : impossible de récupérer la liste",
				Color:       0x831010,
			}
			return
		}
		embeds[id] = &channel.MessageEmbed{
			Title:       str,
			Description: genTopsMessage(tops),
			Color:       0x10E6AD,
		}
	}
	cfg := config.GetGuildConfig(i.GuildID)
	if cfg.DaysXPRemains > 30 {
		wg.Add(1)
		go fn(fmt.Sprintf("Top %d jours", cfg.DaysXPRemains), 10, -1, 0)
	}
	wg.Add(2)
	go fn("Top 30 jours", 5, 30, 1)
	go fn("Top 7 jours", 5, 7, 2)
	go func() {
		wg.Wait()
		if cfg.DaysXPRemains > 30 {
			resp.AddEmbed(embeds[0]).
				AddEmbed(embeds[1]).
				AddEmbed(embeds[2])
		} else {
			resp.AddEmbed(embeds[1]).
				AddEmbed(embeds[2])
		}
		err = resp.Send()
		if err != nil {
			s.Logger().Error("sending response top", "error", err)
		}
	}()
}

func genTopsMessage(tops []user.CopaingAccess) string {
	msg := ""
	for i, c := range tops {
		msg += fmt.Sprintf("%d. **<@%s>** - niveau %d", i+1, c.Copaing().DiscordID, exp.Level(c.GetXP()))
		if i != len(tops)-1 {
			msg += "\n"
		}
	}
	return msg
}
