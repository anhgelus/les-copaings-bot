package user

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/logger"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"slices"
	"sync"
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

func (c *Copaing) AddXP(s *discordgo.Session, m *discordgo.Member, xp uint, fn func(uint, uint)) {
	old, err := c.GetXP()
	if err != nil {
		logger.Alert("user/xp.go - Getting xp", err.Error(), "discord_id", c.DiscordID, "guild_id", c.GuildID)
		return
	}
	pastLevel := exp.Level(old)
	logger.Debug("Adding xp", "member", m.DisplayName(), "old xp", old, "xp to add", xp, "old level", pastLevel)
	c.CopaingXPs = append(c.CopaingXPs, CopaingXP{CopaingID: c.ID, XP: xp, GuildID: c.GuildID})
	if err = c.Save(); err != nil {
		logger.Alert(
			"user/xp.go - Saving user",
			err.Error(),
			"xp",
			c.CopaingXPs,
			"discord_id",
			c.DiscordID,
			"guild_id",
			c.GuildID,
		)
		return
	}
	newLevel := exp.Level(old + xp)
	if newLevel > pastLevel {
		fn(old+xp, newLevel)
		onNewLevel(s, m, newLevel)
	}
}

func (c *Copaing) GetXP() (uint, error) {
	cfg := config.GetGuildConfig(c.GuildID)
	return c.GetXPForDays(cfg.DaysXPRemains)
}

func (c *Copaing) GetXPForDays(n uint) (uint, error) {
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
	defer rows.Close()
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var cxp CopaingXP
		err = gokord.DB.ScanRows(rows, &cxp)
		if err != nil {
			logger.Alert("user/xp.go - Scanning rows", err.Error(), "copaing_id", c.ID, "guild_id", c.GuildID)
			continue
		}
		xp += cxp.XP
	}
	return xp, nil
}

// GetBestXP returns n Copaing with the best XP within d days (d <= cfg.DaysXPRemain; d < 0 <=> d = cfg.DaysXPRemain)
//
// This function is slow
func GetBestXP(guildId string, n uint, d int) ([]CopaingAccess, error) {
	if d < 0 {
		cfg := config.GetGuildConfig(guildId)
		d = int(cfg.DaysXPRemains)
	}
	rows, err := gokord.DB.Model(&Copaing{}).Where("guild_id = ?", guildId).Rows()
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var l []*cXP
	wg := sync.WaitGroup{}
	for rows.Next() {
		var c Copaing
		err = gokord.DB.ScanRows(rows, &c)
		if err != nil {
			logger.Alert("user/xp.go - Scanning rows", err.Error(), "guild_id", guildId)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			xp, err := c.GetXPForDays(uint(d))
			if err != nil {
				logger.Alert("user/xp.go - Fetching xp", err.Error(), "discord_id", c.DiscordID, "guild_id", guildId)
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
