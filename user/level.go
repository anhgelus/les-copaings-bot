package user

import (
	"slices"
	"sync"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/gokord"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/user"
)

func onNewLevel(s *discordgo.Session, m *user.Member, level uint) {
	cfg := config.GetGuildConfig(m.GuildID)
	xpForLevel := exp.LevelXP(level)
	for _, role := range cfg.XpRoles {
		if role.XP <= xpForLevel && !slices.Contains(m.Roles, role.RoleID) {
			s.LogDebug("add role %s to %s in %s", role.RoleID, m.DisplayName(), m.GuildID)
			err := s.GuildAPI().MemberRoleAdd(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				s.LogError(err, "adding role %s to %s in %s", role.RoleID, m.DisplayName(), m.GuildID)
			}
		} else if role.XP > xpForLevel && slices.Contains(m.Roles, role.RoleID) {
			s.LogDebug("remove role %s to %s in %s", role.RoleID, m.DisplayName(), m.GuildID)
			err := s.GuildAPI().MemberRoleRemove(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				s.LogError(err, "removing role s to %s in %s", role.RoleID, m.DisplayName(), m.GuildID)
			}
		}
	}
}

func (c *Copaing) OnNewLevel(s *discordgo.Session, level uint) {
	m, err := s.GuildAPI().Member(c.GuildID, c.DiscordID)
	if err != nil {
		s.LogError(err, "getting member %s in %s for new level", c.DiscordID, c.GuildID)
		return
	}
	onNewLevel(s, m, level)
}

func PeriodicReducer(s *discordgo.Session) {
	wg := &sync.WaitGroup{}
	var cs []*Copaing
	if err := gokord.DB.Find(&cs).Error; err != nil {
		s.LogError(err, "fetching all copaings")
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
			xp, err := c.GetXP()
			if err != nil {
				s.LogError(err, "getting xp of copaing %d in %s", c.ID, c.GuildID)
				xp = 0
			}
			cxps[i] = &cXP{
				Cxp:     xp,
				Copaing: c,
			}
		}()
	}
	wg.Wait()
	for _, g := range s.State.Guilds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := config.GetGuildConfig(g.ID)
			res := gokord.DB.
				Model(&CopaingXP{}).
				Where("guild_id = ? and created_at < ?", g.ID, exp.TimeStampNDaysBefore(cfg.DaysXPRemains)).
				Delete(&CopaingXP{})
			if res.Error != nil {
				s.LogError(res.Error, "removing old xp in %s", g.ID)
			}
			s.LogDebug("Guild cleaned %s, rows affected: %d", g.Name, res.RowsAffected)
		}()
	}
	wg.Wait()
	for i, c := range cxps {
		if i%50 == 49 {
			s.LogDebug("Sleeping...")
			time.Sleep(15 * time.Second) // prevents spamming the API
		}
		oldXp := c.GetXP()
		xp, err := c.ToCopaing().GetXP()
		if err != nil {
			s.LogError(err, "getting xp of copaing %s in %s", c.ID, c.GuildID)
			continue
		}
		if exp.Level(oldXp) != exp.Level(xp) {
			c.OnNewLevel(s, exp.Level(xp))
		}
	}
	s.LogDebug("Periodic reduce finished for %d guilds", len(s.State.Guilds))
}
