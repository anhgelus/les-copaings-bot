package commands

import (
	"bytes"
	"image/color"
	"io"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type data struct {
	CreatedAt time.Time
	XP        int
	CopaingID int
}

type dbData struct {
	CreatedAt *pgtype.Date
	XP        int
	CopaingID int
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
	var w io.WriterTo
	var err error
	if v, ok := opt["user"]; ok {
		w, err = statsMember(s, i, days, v.UserValue(s).ID)
	} else {
		w, err = statsAll(s, i, days)
	}
	if err != nil {
		if err = resp.IsEphemeral().SetMessage("Il y a eu une erreur...").Send(); err != nil {
			logger.Alert("commands/stats.go - Sending error occurred", err.Error())
		}
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

func statsAll(s *discordgo.Session, i *discordgo.InteractionCreate, days uint) (io.WriterTo, error) {
	var rawData []*data
	if gokord.Debug {
		var rawCopaingData []*user.CopaingXP
		err := gokord.DB.
			Where("guild_id = ? and created_at > ?", i.GuildID, exp.TimeStampNDaysBefore(days)).
			Find(&rawCopaingData).
			Error
		if err != nil {
			logger.Alert("commands/stats.go - Fetching result", err.Error())
			return nil, err
		}
		rawData = make([]*data, len(rawCopaingData))
		for in, d := range rawCopaingData {
			rawData[in] = &data{
				CreatedAt: d.CreatedAt,
				XP:        int(d.XP),
				CopaingID: int(d.CopaingID),
			}
		}
	} else {
		sql := `SELECT "created_at"::date::text, sum("xp") as xp, "copaing_id" FROM copaing_xps WHERE "guild_id" = ? and "created_at" > ? GROUP BY "created_at"::date, "copaing_id"`
		var rawDbData []dbData
		res := gokord.DB.Raw(sql, i.GuildID, exp.TimeStampNDaysBefore(days))
		if err := res.Scan(&rawDbData).Error; err != nil {
			logger.Alert("commands/stats.go - Fetching result", err.Error())
			return nil, err
		}
		rawData = make([]*data, len(rawDbData))
		for in, d := range rawDbData {
			rawData[in] = &data{
				CreatedAt: d.CreatedAt.Time,
				XP:        d.XP,
				CopaingID: d.CopaingID,
			}
		}
	}

	copaings := map[int]*user.Copaing{}
	stats := map[int]*[]plotter.XY{}

	for _, raw := range rawData {
		_, ok := copaings[raw.CopaingID]
		if !ok {
			var cp user.Copaing
			if err := gokord.DB.First(&cp, raw.CopaingID).Error; err != nil {
				logger.Alert("commands/stats.go - Finding copaing", err.Error(), "id", raw.CopaingID)
				return nil, err
			}
			copaings[raw.CopaingID] = &cp
		}
		pts, ok := stats[raw.CopaingID]
		if !ok {
			t := make([]plotter.XY, days)
			pts = &t
			stats[raw.CopaingID] = pts
		}
		t := raw.CreatedAt.Unix() - time.Now().Unix()
		if !gokord.Debug {
			t = int64(math.Ceil(float64(t) / (24 * 60 * 60)))
		} else {
			t = int64(math.Ceil(float64(t) / 6))
		}
		(*pts)[t] = plotter.XY{
			X: float64(t),
			Y: float64(raw.XP),
		}
	}

	return generatePlot(s, i, days, copaings, stats)
}

func statsMember(s *discordgo.Session, i *discordgo.InteractionCreate, days uint, discordID string) (io.WriterTo, error) {
	cp := user.GetCopaing(discordID, i.GuildID)
	var rawData []*data
	if gokord.Debug {
		var rawCopaingData []*user.CopaingXP
		err := gokord.DB.
			Where(
				"guild_id = ? and created_at > ? and copaing_id = ?",
				i.GuildID, exp.TimeStampNDaysBefore(days), discordID,
			).
			Find(&rawCopaingData).
			Error
		if err != nil {
			logger.Alert("commands/stats.go - Fetching result", err.Error())
			return nil, err
		}
		rawData = make([]*data, len(rawCopaingData))
		for in, d := range rawCopaingData {
			rawData[in] = &data{
				CreatedAt: d.CreatedAt,
				XP:        int(d.XP),
				CopaingID: int(d.CopaingID),
			}
		}
	} else {
		sql := `SELECT "created_at"::date::text, sum("xp") as xp, "copaing_id" FROM copaing_xps WHERE "guild_id" = ? and "created_at" > ? and "copaing_id" = ? GROUP BY "created_at"::date, "copaing_id"`
		var rawDbData []dbData
		res := gokord.DB.Raw(sql, i.GuildID, exp.TimeStampNDaysBefore(days), discordID)
		if err := res.Scan(&rawDbData).Error; err != nil {
			logger.Alert("commands/stats.go - Fetching result", err.Error())
			return nil, err
		}
		rawData = make([]*data, len(rawDbData))
		for in, d := range rawDbData {
			rawData[in] = &data{
				CreatedAt: d.CreatedAt.Time,
				XP:        d.XP,
				CopaingID: d.CopaingID,
			}
		}
	}

	copaings := map[int]*user.Copaing{
		int(cp.ID): cp,
	}
	stats := map[int]*[]plotter.XY{}

	for _, raw := range rawData {
		pts, ok := stats[raw.CopaingID]
		if !ok {
			t := make([]plotter.XY, days)
			pts = &t
			stats[raw.CopaingID] = pts
		}
		t := raw.CreatedAt.Unix() - time.Now().Unix()
		if !gokord.Debug {
			t = int64(math.Ceil(float64(t) / (24 * 60 * 60)))
		} else {
			t = int64(math.Ceil(float64(t) / 6))
		}
		(*pts)[t] = plotter.XY{
			X: float64(t),
			Y: float64(raw.XP),
		}
	}

	return generatePlot(s, i, days, copaings, stats)
}

func generatePlot(s *discordgo.Session, i *discordgo.InteractionCreate, days uint, copaings map[int]*user.Copaing, stats map[int]*[]plotter.XY) (io.WriterTo, error) {
	p := plot.New()
	p.Title.Text = "Évolution de l'XP"
	p.X.Label.Text = "Jours"
	p.Y.Label.Text = "XP"

	p.Add(plotter.NewGrid())

	r := rand.New(rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().Unix())))
	for in, c := range copaings {
		m, err := s.GuildMember(i.GuildID, c.DiscordID)
		if err != nil {
			logger.Alert("commands/stats.go - Fetching guild member", err.Error())
			return nil, err
		}
		slices.SortFunc(*stats[in], func(a, b plotter.XY) int {
			if a.X < b.X {
				return -1
			}
			if a.X > b.X {
				return 1
			}
			return 0
		})
		first := (*stats[in])[0]
		if first.X > float64(-days) {
			*stats[in] = append([]plotter.XY{{
				X: first.X - 1, Y: 0,
			}}, *stats[in]...)
		}
		last := (*stats[in])[len(*stats[in])-1]
		if last.X <= -1 {
			*stats[in] = append(*stats[in], plotter.XY{
				X: last.X + 1, Y: 0,
			})
		}
		l, err := plotter.NewLine(plotter.XYs(*stats[in]))
		if err != nil {
			logger.Alert("commands/stats.go - Adding line points", err.Error())
			return nil, err
		}
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
		l.LineStyle.Color = color.RGBA{R: uint8(r.UintN(255)), G: uint8(r.UintN(255)), B: uint8(r.UintN(255)), A: 255}
		p.Add(l)
		p.Legend.Add(m.DisplayName(), l)
	}
	w, err := p.WriterTo(8*vg.Inch, 6*vg.Inch, "png")
	if err != nil {
		logger.Alert("commands/stats.go - Generating png", err.Error())
		return nil, err
	}
	return w, nil
}
