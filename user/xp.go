package user

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"slices"
	"sync"
	"time"
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
		utils.SendAlert("user/xp.go - Getting xp", err.Error(), "discord_id", c.DiscordID, "guild_id", c.GuildID)
		return
	}
	pastLevel := exp.Level(old)
	utils.SendDebug("Adding xp", "member", m.DisplayName(), "old xp", old, "xp to add", xp, "old level", pastLevel)
	c.CopaingXPs = append(c.CopaingXPs, CopaingXP{CopaingID: c.ID, XP: xp, GuildID: c.GuildID})
	if err = c.Save(); err != nil {
		utils.SendAlert(
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
	var y, d int
	var m time.Month
	if gokord.Debug {
		y, m, d = time.Unix(time.Now().Unix()-int64(24*60*60), 0).Date() // reduce time for debug
	} else {
		y, m, d = time.Unix(time.Now().Unix()-int64(n*24*60*60), 0).Date()
	}
	rows, err := gokord.DB.
		Model(&CopaingXP{}).
		Where(fmt.Sprintf("created_at >= '%d-%d-%d' and guild_id = ? and copaing_id = ?", y, m, d), c.GuildID, c.ID).
		Rows()
	defer rows.Close()
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var cXP CopaingXP
		err = gokord.DB.ScanRows(rows, &cXP)
		if err != nil {
			utils.SendAlert("user/xp.go - Scanning rows", err.Error(), "copaing_id", c.ID, "guild_id", c.GuildID)
			continue
		}
		xp += cXP.XP
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
			utils.SendAlert("user/xp.go - Scanning rows", err.Error(), "guild_id", guildId)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			xp, err := c.GetXPForDays(uint(d))
			if err != nil {
				utils.SendAlert("user/xp.go - Fetching xp", err.Error(), "discord_id", c.DiscordID, "guild_id", guildId)
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
