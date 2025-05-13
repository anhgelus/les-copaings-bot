package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
	"sync"
)

func Top(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := utils.ResponseBuilder{C: s, I: i}
	err := resp.IsDeferred().Send()
	if err != nil {
		utils.SendAlert("commands/top.go - Sending deferred", err.Error())
		return
	}
	resp.NotDeferred().IsEdit()
	embeds := make([]*discordgo.MessageEmbed, 3)
	wg := sync.WaitGroup{}

	fn := func(s string, n uint, d int, id int) {
		defer wg.Done()
		tops, err := user.GetBestXP(i.GuildID, n, d)
		if err != nil {
			utils.SendAlert("commands/top.go - Fetching best xp", err.Error(), "n", n, "d", d, "id", id, "guild_id", i.GuildID)
			embeds[id] = &discordgo.MessageEmbed{
				Title:       s,
				Description: "Erreur : impossible de récupérer la liste",
				Color:       utils.Error,
			}
			return
		}
		embeds[id] = &discordgo.MessageEmbed{
			Title:       s,
			Description: genTopsMessage(tops),
			Color:       utils.Success,
		}
	}
	cfg := config.GetGuildConfig(i.GuildID)
	if cfg.DaysXPRemains > 30 {
		wg.Add(1)
		go fn("Top full time", 10, -1, 0)
	}
	wg.Add(2)
	go fn("Top 30 jours", 5, 30, 1)
	go fn("Top 7 jours", 5, 7, 2)
	go func() {
		wg.Wait()
		if cfg.DaysXPRemains > 30 {
			resp.Embeds(embeds)
		} else {
			resp.Embeds(embeds[1:])
		}
		err = resp.Send()
		if err != nil {
			utils.SendAlert("commands/top.go - Sending response top", err.Error())
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
