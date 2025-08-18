package user

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"slices"
	"sync"
	"time"
)

func onNewLevel(dg *discordgo.Session, m *discordgo.Member, level uint) {
	cfg := config.GetGuildConfig(m.GuildID)
	xpForLevel := exp.LevelXP(level)
	for _, role := range cfg.XpRoles {
		if role.XP <= xpForLevel && !slices.Contains(m.Roles, role.RoleID) {
			utils.SendDebug(
				"Add role",
				"role_id", role.RoleID,
				"user_id", m.User.ID,
				"guild_id", m.GuildID,
			)
			err := dg.GuildMemberRoleAdd(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				utils.SendAlert("user/level.go - Adding role", err.Error(), "role_id", role.RoleID)
			}
		} else if role.XP > xpForLevel && slices.Contains(m.Roles, role.RoleID) {
			utils.SendDebug(
				"Remove role",
				"role_id", role.RoleID,
				"user_id", m.User.ID,
				"guild_id", m.GuildID,
			)
			err := dg.GuildMemberRoleRemove(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				utils.SendAlert("user/level.go - Removing role", err.Error(), "role_id", role.RoleID)
			}
		}
	}
}

func (c *Copaing) OnNewLevel(dg *discordgo.Session, level uint) {
	m, err := dg.GuildMember(c.GuildID, c.DiscordID)
	if err != nil {
		utils.SendAlert(
			"user/level.go - Getting member for new level", err.Error(),
			"discord_id", c.DiscordID,
			"guild_id", c.GuildID,
		)
		return
	}
	onNewLevel(dg, m, level)
}

func PeriodicReducer(dg *discordgo.Session) {
	wg := &sync.WaitGroup{}
	var cs []*Copaing
	if err := gokord.DB.Find(&cs).Error; err != nil {
		utils.SendAlert("user/level.go - Fetching all copaings", err.Error())
		return
	}
	cxps := make([]*cXP, len(cs))
	for i, c := range cs {
		if i%10 == 9 {
			wg.Wait() // prevents spamming the DB
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			xp, err := c.GetXP()
			if err != nil {
				utils.SendAlert("user/level.go - Getting XP", err.Error(), "copaing_id", c.ID, "guild_id", c.GuildID)
				xp = 0
			}
			cxps[i] = &cXP{
				Cxp:     xp,
				Copaing: c,
			}
		}()
	}
	wg.Wait()
	for _, g := range dg.State.Guilds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := config.GetGuildConfig(g.ID)
			res := gokord.DB.
				Model(&CopaingXP{}).
				Where("guild_id = ? and created_at < ?", g.ID, exp.TimeStampNDaysBefore(cfg.DaysXPRemains)).
				Delete(&CopaingXP{})
			if res.Error != nil {
				utils.SendAlert("user/level.go - Removing old XP", res.Error.Error(), "guild_id", g.ID)
			}
			utils.SendDebug("Guild cleaned", "guild", g.Name, "rows affected", res.RowsAffected)
		}()
	}
	wg.Wait()
	for i, c := range cxps {
		if i%50 == 49 {
			utils.SendDebug("Sleeping...")
			time.Sleep(15 * time.Second) // prevents spamming the API
		}
		oldXp := c.GetXP()
		xp, err := c.ToCopaing().GetXP()
		if err != nil {
			utils.SendAlert("user/level.go - Getting XP", err.Error(), "guild_id", c.ID, "discord_id", c.DiscordID)
			continue
		}
		if exp.Level(oldXp) != exp.Level(xp) {
			c.OnNewLevel(dg, exp.Level(xp))
		}
	}
	utils.SendDebug("Periodic reduce finished", "len(guilds)", len(dg.State.Guilds))
}
