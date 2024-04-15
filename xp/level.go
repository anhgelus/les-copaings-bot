package xp

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/bwmarrin/discordgo"
	"slices"
)

func onNewLevel(s *discordgo.Session, m *discordgo.Member, level uint) {
	cfg := config.GetGuildConfig(m.GuildID)
	xpForLevel := XPForLevel(level)
	for _, role := range cfg.XpRoles {
		if role.XP <= xpForLevel && !slices.Contains(m.Roles, role.RoleID) {
			utils.SendDebug(
				"Add role",
				"role_id",
				role.RoleID,
				"user_id",
				m.User.ID,
				"guild_id",
				m.GuildID,
			)
			err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				utils.SendAlert("xp/level.go - Adding role", err.Error(), "role_id", role.RoleID)
			}
		} else if role.XP > xpForLevel && slices.Contains(m.Roles, role.RoleID) {
			utils.SendDebug(
				"Remove role",
				"role_id",
				role.RoleID,
				"user_id",
				m.User.ID,
				"guild_id",
				m.GuildID,
			)
			err := s.GuildMemberRoleRemove(m.GuildID, m.User.ID, role.RoleID)
			if err != nil {
				utils.SendAlert("xp/level.go - Removing role", err.Error(), "role_id", role.RoleID)
			}
		}
	}
}

func LastEventUpdate(c *Copaing) {
	h := c.HourSinceLastEvent()
	l := Lose(h, c.XP)
	xp := c.XPAlreadyRemoved()
	if l-xp < 0 {
		utils.SendWarn("lose - xp already removed is negative", "lose", l, "xp", xp)
		c.XP = 0
	} else {
		c.XP = c.XP - l + c.XPAlreadyRemoved()
	}
	c.Save()
	c.SetLastEvent()
}

func XPUpdate(c *Copaing) {
	h := c.HourSinceLastEvent()
	l := Lose(h, c.XP)
	xp := c.XPAlreadyRemoved()
	if l-xp < 0 {
		utils.SendWarn("lose - xp already removed is negative", "lose", l, "xp", xp)
		c.AddXPAlreadyRemoved(0)
	} else {
		c.XP = c.XP - l + xp
		c.AddXPAlreadyRemoved(l - xp)
	}
	c.Save()
}
