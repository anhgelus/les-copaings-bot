package commands

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"slices"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/jackc/pgx/v5/pgtype"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/channel"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gorm.io/gorm"
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

var colors = []color.RGBA{
	{38, 70, 83, 255},
	{42, 157, 143, 255},
	{244, 162, 97, 255},
	{231, 111, 81, 255},
	{193, 18, 31, 255},
}

func Stats(s *discordgo.Session, i *discordgo.InteractionCreate, opt cmd.OptionMap, resp *cmd.ResponseBuilder) {
	cfg := config.GetGuildConfig(i.GuildID)
	days := 15
	if gokord.Debug {
		days = 90
	}
	if v, ok := opt["days"]; ok {
		in := v.IntValue()
		if in < 1 || uint(in) > cfg.DaysXPRemains {
			msg := fmt.Sprintf("Nombre de jours invalide. Il doit être strictement positif et inférieur à %d", cfg.DaysXPRemains)
			if err := resp.SetMessage(msg).IsEphemeral().Send(); err != nil {
				s.LogError(err, "sending error invalid days")
			}
			return
		}
		days = int(in)
	}
	err := resp.IsDeferred().Send()
	if err != nil {
		s.LogError(err, "sending deferred")
		return
	}
	go func() {
		var w io.WriterTo
		if v, ok := opt["user"]; ok {
			w, err = statsMember(s, i, days, v.UserValue(s.UserAPI()).ID)
		} else {
			w, err = statsAll(s, i, days)
		}
		if err != nil {
			s.LogError(err, "generating stats in %s", i.GuildID)
			if err = resp.IsEphemeral().SetMessage("Il y a eu une erreur...").Send(); err != nil {
				s.LogError(err, "sending error occurred")
			}
			return
		}
		b := new(bytes.Buffer)
		_, err = w.WriteTo(b)
		if err != nil {
			s.LogError(err, "writing png")
		}
		err = resp.AddFile(&channel.File{
			Name:        "plot.png",
			ContentType: "image/png",
			Reader:      b,
		}).Send()
		if err != nil {
			s.LogError(err, "sending stats")
		}
	}()
}

func statsAll(s *discordgo.Session, i *discordgo.InteractionCreate, days int) (io.WriterTo, error) {
	return stats(s, i, days, func(before, after string) *gorm.DB {
		return gokord.DB.Raw(before+"WHERE guild_id = ? and created_at > ?"+after, i.GuildID, exp.TimeStampNDaysBefore(uint(days)))
	})
}

func statsMember(s *discordgo.Session, i *discordgo.InteractionCreate, days int, discordID string) (io.WriterTo, error) {
	_, err := s.GuildAPI().Member(i.GuildID, discordID)
	if err != nil {
		return nil, err
	}
	return stats(s, i, days, func(before, after string) *gorm.DB {
		return gokord.DB.Raw(
			before+"WHERE guild_id = ? and created_at > ? and copaing_id = ?"+after,
			i.GuildID, exp.TimeStampNDaysBefore(uint(days)), user.GetCopaing(discordID, i.GuildID).ID,
		)
	})
}

func stats(s *discordgo.Session, i *discordgo.InteractionCreate, days int, execSql func(before, after string) *gorm.DB) (io.WriterTo, error) {
	var rawData []*data
	if gokord.Debug {
		var rawCopaingData []*user.CopaingXP
		if err := execSql("SELECT * FROM copaing_xps ", "").Scan(&rawCopaingData).Error; err != nil {
			s.LogError(err, "fetching result")
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
		var rawDbData []dbData
		if err := execSql(
			`SELECT "created_at"::date::text, sum("xp") as xp, "copaing_id" FROM copaing_xps `, ` GROUP BY "created_at"::date, "copaing_id"`,
		).Scan(&rawDbData).Error; err != nil {
			s.LogError(err, "fetching result")
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
	stats := map[int][]plotter.XY{}

	for _, raw := range rawData {
		_, ok := copaings[raw.CopaingID]
		if !ok {
			var cp user.Copaing
			if err := gokord.DB.First(&cp, raw.CopaingID).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					s.LogError(err, "finding copaing %d", raw.CopaingID)
					return nil, err
				}
				s.LogWarn("Copaing %d not found, skipping", raw.CopaingID)
				continue
			}
			copaings[raw.CopaingID] = &cp
		}
		pts, ok := stats[raw.CopaingID]
		now := time.Now().Unix()
		if !ok {
			pts = make([]plotter.XY, days+1)
			for i := 0; i < len(pts); i++ {
				pts[i] = plotter.XY{
					X: float64(i - int(days)),
					Y: 0,
				}
			}
			stats[raw.CopaingID] = pts
		}
		t := raw.CreatedAt.Unix() - now
		if !gokord.Debug {
			t = int64(math.Ceil(float64(t) / (24 * 60 * 60)))
		}
		pts[int64(days)+t] = plotter.XY{ // because t <= 0
			X: float64(t),
			Y: float64(raw.XP),
		}
	}
	return generatePlot(s, i, copaings, stats)
}

func generatePlot(s *discordgo.Session, i *discordgo.InteractionCreate, copaings map[int]*user.Copaing, stats map[int][]plotter.XY) (io.WriterTo, error) {
	p := plot.New()
	p.Title.Text = "Évolution de l'XP"
	p.X.Label.Text = "Jours"
	if gokord.Debug {
		p.X.Label.Text = "Secondes"
	}
	p.Y.Label.Text = "XP"
	p.Y.Scale = exp.LevelScale{}

	p.Add(plotter.NewGrid())

	cnt := 0
	for in, c := range copaings {
		m, err := s.GuildAPI().Member(i.GuildID, c.DiscordID)
		if err != nil {
			s.LogError(err, "fetching guild member")
			return nil, err
		}
		slices.SortFunc(stats[in], func(a, b plotter.XY) int {
			if a.X < b.X {
				return -1
			}
			if a.X > b.X {
				return 1
			}
			return 0
		})
		l, _, err := plotter.NewLinePoints(plotter.XYs(stats[in]))
		if err != nil {
			s.LogError(err, "adding line points")
			return nil, err
		}
		l.Color = colors[cnt%len(colors)]
		p.Add(l)
		p.Legend.Add(m.DisplayName(), l)
		cnt++
	}
	w, err := p.WriterTo(8*vg.Inch, 6*vg.Inch, "png")
	if err != nil {
		s.LogError(err, "generating png")
		return nil, err
	}
	return w, nil
}
