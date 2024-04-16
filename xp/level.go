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
	oldXP := c.XP
	if l-xp < 0 {
		utils.SendWarn("lose - xp already removed is negative", "lose", l, "xp", xp)
		c.XP = 0
	} else {
		calc := int(c.XP) - int(l) + int(c.XPAlreadyRemoved())
		if calc < 0 {
			c.XP = 0
		} else {
			c.XP = uint(calc)
		}
	}
	if oldXP != c.XP {
		c.Save()
	}
	c.SetLastEvent()
}

func XPUpdate(c *Copaing) {
	oldXP := c.XP
	if oldXP == 0 {
		return
	}
	h := c.HourSinceLastEvent()
	l := Lose(h, c.XP)
	xp := c.XPAlreadyRemoved()
	if l-xp < 0 {
		utils.SendWarn("lose - xp_removed is negative", "lose", l, "xp removed", xp)
		c.AddXPAlreadyRemoved(0)
	} else {
		calc := int(c.XP) - int(l) + int(xp)
		if calc < 0 {
			c.AddXPAlreadyRemoved(c.XP)
			c.XP = 0
		} else {
			c.XP = uint(calc)
			c.AddXPAlreadyRemoved(l - xp)
		}
	}
	if oldXP != c.XP {
		utils.SendDebug("Save XP", "old", oldXP, "new", c.XP, "user", c.DiscordID)
		c.Save()
	}
}

//func PeriodicReducer(s *discordgo.Session) {
//	var wg sync.WaitGroup
//	for _, g := range s.State.Guilds {
//		for _, m := range utils.FetchGuildUser(s, g.ID) {
//			if m.User.Bot {
//				continue
//			}
//			wg.Add(1)
//			go func() {
//				utils.SendDebug("Async reducer", "user", m.DisplayName(), "guild", g.Name)
//				c := GetCopaing(m.User.ID, g.ID)
//				XPUpdate(c)
//				wg.Done()
//			}()
//		}
//		wg.Wait() // finish the entire guild before starting another
//		utils.SendDebug("Guild finished", "guild", g.Name)
//		time.Sleep(10 * time.Second) // sleep prevents from spamming the Discord API
//	}
//	utils.SendDebug("Periodic reduce finished", "len(guilds)", len(s.State.Guilds))
//}
