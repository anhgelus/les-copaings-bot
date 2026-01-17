package user

import (
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

func (c *Copaing) OnNewLevel(s *discordgo.Session, level uint) {
	m, err := s.GuildAPI().Member(c.GuildID, c.DiscordID)
	if err != nil {
		s.Logger().Error("getting member for new level", "error", err, "user", c.DiscordID, "guild", c.GuildID)
		return
	}
	onNewLevel(s, m, level)
}

func PeriodicReducer(s *discordgo.Session) {
	wg := &sync.WaitGroup{}
	var cs []*Copaing
	if err := gokord.DB.Find(&cs).Error; err != nil {
		s.Logger().Error("fetching all copaings", "error", err)
		return
	}
	cxps := make([]*cXP, len(cs))
	for i, c := range cs {
		if i%25 == 24 {
			wg.Wait() // prevents spamming the DB
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			xp, err := c.GetXP(s.Logger())
			if err != nil {
				s.Logger().Error("getting xp", "error", err, "copaing", c.ID, "guild", c.GuildID)
				xp = 0
			}
			cxps[i] = &cXP{
				Cxp:     xp,
				copaing: c,
			}
		}()
	}
	wg.Wait()
	i := 0
	for _, g := range s.GuildAPI().State.Guilds() {
		i++
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := config.GetGuildConfig(g)
			res := gokord.DB.
				Model(&CopaingXP{}).
				Where("guild_id = ? and created_at < ?", g, exp.TimeStampNDaysBefore(cfg.DaysXPRemains)).
				Delete(&CopaingXP{})
			if res.Error != nil {
				s.Logger().Error("removing old xp", "error", res.Error, "guild", g)
			}
			s.Logger().Debug("guild cleaned", "guild", g, "rows affected", res.RowsAffected)
		}()
	}
	wg.Wait()
	for i, c := range cxps {
		if i%50 == 49 {
			s.Logger().Debug("sleeping...")
			time.Sleep(15 * time.Second) // prevents spamming the API
		}
		oldXp := c.GetXP()
		cp := c.Copaing()
		xp, err := cp.GetXP(s.Logger())
		if err != nil {
			s.Logger().Error("getting xp of copaing", "error", err, "copaing", cp.ID, "guild", cp.GuildID)
			continue
		}
		if exp.Level(oldXp) != exp.Level(xp) {
			cp.OnNewLevel(s, exp.Level(xp))
		}
	}
	s.Logger().Debug("periodic reduce finished", "guilds affected", i)
}
