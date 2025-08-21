package commands

import (
	"fmt"
	"sync"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
)

func Top(_ *discordgo.Session, i *discordgo.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	err := resp.IsDeferred().Send()
	if err != nil {
		logger.Alert("commands/top.go - Sending deferred", err.Error())
		return
	}
	embeds := make([]*discordgo.MessageEmbed, 3)
	wg := sync.WaitGroup{}

	fn := func(s string, n uint, d int, id int) {
		defer wg.Done()
		tops, err := user.GetBestXP(i.GuildID, n, d)
		if err != nil {
			logger.Alert("commands/top.go - Fetching best xp", err.Error(), "n", n, "d", d, "id", id, "guild_id", i.GuildID)
			embeds[id] = &discordgo.MessageEmbed{
				Title:       s,
				Description: "Erreur : impossible de récupérer la liste",
				Color:       0x831010,
			}
			return
		}
		embeds[id] = &discordgo.MessageEmbed{
			Title:       s,
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
			logger.Alert("commands/top.go - Sending response top", err.Error())
		}
	}()
}

func genTopsMessage(tops []user.CopaingAccess) string {
	msg := ""
	for i, c := range tops {
		msg += fmt.Sprintf("%d. **<@%s>** - niveau %d", i+1, c.ToCopaing().DiscordID, exp.Level(c.GetXP()))
		if i != len(tops)-1 {
			msg += "\n"
		}
	}
	return msg
}
