package user

import (
	"context"
	"slices"
	"sync"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/common"
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/guild"
	"github.com/nyttikord/gokord/user"
)

func onNewLevel(ctx context.Context, dg bot.Session, m *user.Member, level uint) {
	cfg := config.GetGuildConfig(ctx, m.GuildID)
	xpForLevel := exp.LevelXP(level)
	for _, role := range cfg.XpRoles {
		if role.XP <= xpForLevel && !slices.Contains(m.Roles, role.RoleID) {
			bot.Logger(ctx).Debug("add role", "role", role.RoleID, "user", m.DisplayName(), "guild", m.GuildID)
			err := guild.AddMemberRole(m.GuildID, m.User.ID, role.RoleID).Do(ctx)
			if err != nil {
				bot.Logger(ctx).Error(
					"adding role",
					"error", err, "role", role.RoleID, "user", m.DisplayName(), "guild", m.GuildID,
				)
			}
		} else if role.XP > xpForLevel && slices.Contains(m.Roles, role.RoleID) {
			bot.Logger(ctx).Debug("remove role", "role", role.RoleID, "user", m.DisplayName(), "guild", m.GuildID)
			err := guild.RemoveMemberRole(m.GuildID, m.User.ID, role.RoleID).Do(ctx)
			if err != nil {
				bot.Logger(ctx).Error(
					"removing role",
					"error", err, "role", role.RoleID, "user", m.DisplayName(), "guild", m.GuildID,
				)
			}
		}
	}
}

func (c *CopaingCached) onNewLevel(ctx context.Context, dg bot.Session, level uint) {
	m, err := guild.GetMember(c.GuildID, c.DiscordID).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("getting member for new level", "error", err, "user", c.DiscordID, "guild", c.GuildID)
		return
	}
	onNewLevel(ctx, dg, m, level)
}

func PeriodicReducer(ctx context.Context, dg bot.Session) {
	PeriodicSaver(ctx, dg)

	log := bot.Logger(ctx)
	log.Debug("periodic reducer")

	state := GetState(ctx)

	n := 0
	var wg sync.WaitGroup
	for _, g := range dg.GuildState().ListGuilds() {
		cfg := config.GetGuildConfig(ctx, g)
		res := common.GetDB(ctx).
			Model(&CopaingXP{}).
			Where("guild_id = ? and created_at < ?", g, exp.TimeStampNDaysBefore(ctx, cfg.DaysXPRemains)).
			Delete(&CopaingXP{})
		if res.Error != nil {
			log.Error("removing old xp", "error", res.Error, "guild", g)
			continue
		}
		n++
		log.Debug("guild cleaned", "guild", g, "rows affected", res.RowsAffected)

		wg.Go(func() {
			syncCopaings(ctx, dg, state.Copaings(g))
		})
	}

	wg.Wait()

	log.Debug("periodic reduce finished", "guilds affected", n)
}

func syncCopaings(ctx context.Context, dg bot.Session, ccs []CopaingCached) {
	logger := bot.Logger(ctx)
	for i, cc := range ccs {
		if i%50 == 49 {
			logger.Debug("sleeping...")
			time.Sleep(15 * time.Second) // prevents spamming the API
		}
		oldXp := cc.XP
		err := cc.Sync(ctx)
		if err != nil {
			logger.Error("syncing copaing", "error", err, "copaing", cc.ID, "guild", cc.GuildID)
			continue
		}
		xp := cc.XP
		if exp.Level(oldXp) != exp.Level(xp) {
			cc.onNewLevel(ctx, dg, exp.Level(xp))
		}
	}
}

func PeriodicSaver(ctx context.Context, dg bot.Session) {
	logger := bot.Logger(ctx)
	logger.Debug("saving state in DB")
	err := saveStateInDB(ctx)
	if err != nil {
		panic(err)
	}
}
