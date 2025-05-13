package user

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"slices"
	"sync"
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
	for _, g := range dg.State.Guilds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := config.GetGuildConfig(g.ID)
			err := gokord.DB.
				Model(&CopaingXP{}).
				Where("guild_id = ? and created_at < ?", g.ID, exp.TimeStampNDaysBefore(cfg.DaysXPRemains)).
				Delete(&CopaingXP{}).
				Error
			if err != nil {
				utils.SendAlert("user/level.go - Removing old XP", err.Error(), "guild_id", g.ID)
			}
		}()
	}
	wg.Wait()
	utils.SendDebug("Periodic reduce finished", "len(guilds)", len(dg.State.Guilds))
}
