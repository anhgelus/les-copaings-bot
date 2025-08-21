package commands

import (
	"bytes"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type data struct {
	CreatedAt time.Time
	XP        int
	CopaingID int
	copaing   *user.Copaing
}

func Stats(s *discordgo.Session, i *discordgo.InteractionCreate, opt cmd.OptionMap, resp *cmd.ResponseBuilder) {
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
	var rawData []*data
	res := gokord.DB.Raw(
		`SELECT "created_at"::date::text, sum(xp) as xp, copaing_id FROM copaing_xps GROUP BY "created_at"::date `+
			` WHERE guild_id = ? and created_at < ?`,
		i.GuildID, exp.TimeStampNDaysBefore(days),
	)
	if res.Error != nil {
		logger.Alert("commands/stats.go - Fetching XP data", res.Error.Error(), "guild_id", i.GuildID)
		return
	}
	if err := res.Scan(&rawData).Error; err != nil {
		logger.Alert("commands/stats.go - Scanning result", err.Error(), "res")
		return
	}

	copaings := map[int]*user.Copaing{}
	stats := map[int]*[]*plotter.XY{}

	for _, s := range rawData {
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
		pts, ok := stats[s.CopaingID]
		if !ok {
			pts = &[]*plotter.XY{}
			stats[s.CopaingID] = pts
		}
		t := float64(s.CreatedAt.Unix()-time.Now().Unix()) / (24 * 60 * 60)
		*pts = append(*pts, &plotter.XY{
			X: t,
			Y: float64(s.XP),
		})
	}

	p := plot.New()
	p.Title.Text = "XP"
	p.X.Label.Text = "Jours"
	p.Y.Label.Text = "XP"

	for in, c := range copaings {
		m, err := s.GuildMember(i.GuildID, c.DiscordID)
		if err != nil {
			logger.Alert("commands/stats.go - Fetching guild member", err.Error())
			return
		}
		err = plotutil.AddLinePoints(p, m.DisplayName(), stats[in])
		if err != nil {
			logger.Alert("commands/stats.go - Adding line points", err.Error())
			return
		}
	}
	w, err := p.WriterTo(4*vg.Inch, 4*vg.Inch, "png")
	if err != nil {
		logger.Alert("commands/stats.go - Generating png", err.Error())
		return
	}
	b := new(bytes.Buffer)
	_, err = w.WriteTo(b)
	if err != nil {
		logger.Alert("commands/stats.go - Writing png", err.Error())
	}
	err = resp.AddFile(&discordgo.File{
		Name:        "plot.png",
		ContentType: "image/png",
		Reader:      b,
	}).Send()
	if err != nil {
		logger.Alert("commands/stats.go - Sending response", err.Error())
	}
}
