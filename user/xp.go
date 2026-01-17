package user

import (
	"context"
	"log/slog"
	"slices"
	"sync"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/user"
)

type cXP struct {
	Cxp uint
	*Copaing
}

func (c *cXP) ToCopaing() *Copaing {
	return c.Copaing
}

func (c *cXP) GetXP() uint {
	return c.Cxp
}

func (cc *CopaingCached) AddXP(ctx context.Context, s bot.Session, m *user.Member, xp uint, fn func(uint, uint)) {
	old := cc.XPs
	pastLevel := exp.Level(old)
	s.Logger().Debug("adding xp", "user", m.DisplayName(), "old", old, "to add", xp)
	cc.XPs += xp
	cc.XPToAdd += xp
	if err := cc.Save(ctx); err != nil {
		s.Logger().Error("saving user in state", "error", err, "user", m.DisplayName(), "xp", xp, "guild", cc.GuildID)
		return
	}
	newLevel := exp.Level(old + xp)
	if newLevel > pastLevel {
		fn(old+xp, newLevel)
		onNewLevel(s, m, newLevel)
	}
}

func (c *Copaing) GetXP(logger *slog.Logger) (uint, error) {
	cfg := config.GetGuildConfig(c.GuildID)
	return c.GetXPForDays(logger, cfg.DaysXPRemains)
}

func (c *Copaing) GetXPForDays(logger *slog.Logger, n uint) (uint, error) {
	xp := uint(0)
	rows, err := gokord.DB.
		Model(&CopaingXP{}).
		Where(
			"created_at >= ? and guild_id = ? and copaing_id = ?",
			exp.TimeStampNDaysBefore(n),
			c.GuildID,
			c.ID,
		).
		Rows()
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var cxp CopaingXP
		err = gokord.DB.ScanRows(rows, &cxp)
		if err != nil {
			logger.Error("scanning rows", "error", err, "copaing", c.ID, "guild", c.GuildID)
			continue
		}
		xp += cxp.XP
	}
	return xp, nil
}

// GetBestXP returns n Copaing with the best XP within d days (d <= cfg.DaysXPRemain; d < 0 <=> d = cfg.DaysXPRemain)
//
// This function is slow
func GetBestXP(logger *slog.Logger, guildId string, n uint, d int) ([]CopaingAccess, error) {
	if d < 0 {
		cfg := config.GetGuildConfig(guildId)
		d = int(cfg.DaysXPRemains)
	}
	rows, err := gokord.DB.Model(&Copaing{}).Where("guild_id = ?", guildId).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var l []*cXP
	var wg sync.WaitGroup
	for rows.Next() {
		var c Copaing
		err = gokord.DB.ScanRows(rows, &c)
		if err != nil {
			logger.Error("scanning rows", "error", err, "copaing", c.ID, "guild", c.GuildID)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			xp, err := c.GetXPForDays(logger, uint(d))
			if err != nil {
				logger.Error("fetching xp", "error", err, "copaing", c.ID, "guild", c.GuildID)
				return
			}
			l = append(l, &cXP{Cxp: xp, Copaing: &c})
		}()
	}
	wg.Wait()
	slices.SortFunc(l, func(a, b *cXP) int {
		// desc order
		if a.Cxp < b.Cxp {
			return 1
		}
		if a.Cxp > b.Cxp {
			return -1
		}
		return 0
	})
	m := min(len(l), int(n))
	cs := make([]CopaingAccess, m)
	for i, c := range l[:m] {
		cs[i] = c
	}
	return cs, nil
}
