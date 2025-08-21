package commands

import (
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
)

type data struct {
	CreatedAt time.Time
	XP        int
	CopaingID int
	copaing   *user.Copaing
}

func Stats(_ *discordgo.Session, i *discordgo.InteractionCreate, opt cmd.OptionMap, resp *cmd.ResponseBuilder) {
	cfg := config.GetGuildConfig(i.GuildID)
	days := cfg.DaysXPRemains
	if v, ok := opt["days"]; ok {
		in := v.IntValue()
		if in < 0 || uint(in) > days {
			if err := resp.SetMessage("Nombre de jours invalide").IsEphemeral().Send(); err != nil {
				logger.Alert("commands/stats.go - Sending invalid days", err.Error())
			}
			return
		}
		days = uint(in)
	}
	var stats []*data
	res := gokord.DB.Raw(
		`SELECT "created_at"::date::text, sum(xp) as xp, copaing_id FROM copaing_xps GROUP BY "created_at"::date `+
			` WHERE guild_id = ? and created_at < ?`,
		i.GuildID, exp.TimeStampNDaysBefore(days),
	)
	if res.Error != nil {
		logger.Alert("commands/stats.go - Fetching XP data", res.Error.Error(), "guild_id", i.GuildID)
		return
	}
	if err := res.Scan(&stats).Error; err != nil {
		logger.Alert("commands/stats.go - Scanning result", err.Error(), "res")
		return
	}
	copaings := map[int]*user.Copaing{}
	for _, s := range stats {
		c, ok := copaings[s.CopaingID]
		if ok {
			s.copaing = c
		} else {
			if err := gokord.DB.First(s.copaing, s.CopaingID).Error; err != nil {
				logger.Alert("commands/stats.go - Finding copaing", err.Error(), "id", s.CopaingID)
				return
			}
			copaings[s.CopaingID] = s.copaing
		}
	}
}
