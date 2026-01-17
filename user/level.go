package user

import (
	"context"
	"slices"
	"sync"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/gokord"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/user"
)

func onNewLevel(s bot.Session, m *user.Member, level uint) {
	cfg := config.GetGuildConfig(m.GuildID)
	xpForLevel := exp.LevelXP(level)
	for _, role := range cfg.XpRoles {
		if role.XP <= xpForLevel && !slices.Contains(m.Roles, role.RoleID) {
			s.Logger().Debug("add role", "role", role.RoleID, "user", m.DisplayName(), "guild", m.GuildID)
			err := s.GuildAPI().MemberRoleAdd(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				s.Logger().Error(
					"adding role",
					"error", err,
					"role", role.RoleID,
					"user", m.DisplayName(),
					"guild", m.GuildID,
				)
			}
		} else if role.XP > xpForLevel && slices.Contains(m.Roles, role.RoleID) {
			s.Logger().Debug("remove role", "role", role.RoleID, "user", m.DisplayName(), "guild", m.GuildID)
			err := s.GuildAPI().MemberRoleRemove(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				s.Logger().Error(
					"removing role",
					"error", err,
					"role", role.RoleID,
					"user", m.DisplayName(),
					"guild", m.GuildID,
				)
			}
		}
	}
}

func (c *CopaingCached) onNewLevel(s *discordgo.Session, level uint) {
	m, err := s.GuildAPI().Member(c.GuildID, c.DiscordID)
	if err != nil {
		s.Logger().Error("getting member for new level", "error", err, "user", c.DiscordID, "guild", c.GuildID)
		return
	}
	onNewLevel(s, m, level)
}

func PeriodicReducer(ctx context.Context, s *discordgo.Session) {
	PeriodicSaver(ctx, s)

	s.Logger().Debug("periodic reducer")

	state := GetState(ctx)

	n := 0
	var wg sync.WaitGroup
	for _, g := range s.GuildAPI().State.Guilds() {
		n++
		cfg := config.GetGuildConfig(g)
		res := gokord.DB.
			Model(&CopaingXP{}).
			Where("guild_id = ? and created_at < ?", g, exp.TimeStampNDaysBefore(cfg.DaysXPRemains)).
			Delete(&CopaingXP{})
		if res.Error != nil {
			s.Logger().Error("removing old xp", "error", res.Error, "guild", g)
			continue
		}
		s.Logger().Debug("guild cleaned", "guild", g, "rows affected", res.RowsAffected)

		wg.Go(func() {
			syncCopaings(ctx, s, state.Copaings(g))
		})
	}

	wg.Wait()

	s.Logger().Debug("periodic reduce finished", "guilds affected", n)
}

func syncCopaings(ctx context.Context, s *discordgo.Session, ccs []CopaingCached) {
	for i, cc := range ccs {
		if i%50 == 49 {
			s.Logger().Debug("sleeping...")
			time.Sleep(15 * time.Second) // prevents spamming the API
		}
		oldXp := cc.XP
		err := cc.Sync(ctx)
		if err != nil {
			s.Logger().Error("syncing copaing", "error", err, "copaing", cc.ID, "guild", cc.GuildID)
			continue
		}
		xp := cc.XP
		if exp.Level(oldXp) != exp.Level(xp) {
			cc.onNewLevel(s, exp.Level(xp))
		}
	}
}

func PeriodicSaver(ctx context.Context, s bot.Session) {
	s.Logger().Debug("saving state in DB")
	err := saveStateInDB(ctx)
	if err != nil {
		panic(err)
	}
}
