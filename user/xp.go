package user

import (
	"context"
	"slices"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/user"
)

type cXP struct {
	Cxp     uint
	copaing *Copaing
}

func (c *cXP) Copaing() *Copaing {
	return c.copaing
}

func (c *cXP) GetXP() uint {
	return c.Cxp
}

func (cc *CopaingCached) AddXP(ctx context.Context, s bot.Session, m *user.Member, xp uint, fn func(uint, uint)) {
	old := cc.XP
	pastLevel := exp.Level(old)
	s.Logger().Debug("adding xp", "user", m.DisplayName(), "old", old, "to add", xp)
	cc.XP += xp
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

func (cc *CopaingCached) GetXPForDays(n uint) uint {
	xp := uint(0)
	for _, v := range cc.XPs {
		if v.Time <= time.Duration(n*24)*time.Hour {
			xp += v.XP
		}
	}
	return xp + cc.XPToAdd
}

// GetBestXP returns n Copaing with the best XP within d days (d <= cfg.DaysXPRemain; d < 0 <=> d = cfg.DaysXPRemain)
func GetBestXP(ctx context.Context, guildId string, n uint, d int) []CopaingCached {
	ccs := GetState(ctx).Copaings(guildId)
	if d > 0 {
		for _, v := range ccs {
			v.XP = v.GetXPForDays(n)
		}
	}
	slices.SortFunc(ccs, func(a, b CopaingCached) int {
		// desc order
		return int(b.XP) - int(a.XP)
	})
	m := min(len(ccs), int(n))
	res := make([]CopaingCached, m)
	copy(ccs[:m], res)
	return res
}
