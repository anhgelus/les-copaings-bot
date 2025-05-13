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
	var wg sync.WaitGroup
	for _, g := range dg.State.Guilds {
		var cs []*Copaing
		err := gokord.DB.Where("guild_id = ?", g.ID).Find(&cs).Error
		if err != nil {
			utils.SendAlert("user/level.go - Querying all copaings in Guild", err.Error(), "guild_id", g.ID)
			continue
		}
		for i, c := range cs {
			if i%50 == 49 {
				time.Sleep(15 * time.Second) // sleep prevents from spamming the Discord API and the database
			}
			var u *discordgo.User
			u, err = dg.User(c.DiscordID)
			if err != nil {
				utils.SendAlert(
					"user/level.go - Fetching user", err.Error(),
					"discord_id", c.DiscordID,
					"guild_id", g.ID,
				)
				utils.SendWarn("Removing user from database", "discord_id", c.DiscordID)
				if err = gokord.DB.Delete(c).Error; err != nil {
					utils.SendAlert(
						"user/level.go - Removing user from database", err.Error(),
						"discord_id", c.DiscordID,
						"guild_id", g.ID,
					)
				}
				continue
			}
			if u.Bot {
				continue
			}
			if _, err = dg.GuildMember(g.ID, c.DiscordID); err != nil {
				utils.SendAlert(
					"user/level.go - Fetching member", err.Error(),
					"discord_id", c.DiscordID,
					"guild_id", g.ID,
				)
				utils.SendWarn(
					"Removing user from guild in database",
					"discord_id", c.DiscordID,
					"guild_id", g.ID,
				)
				if err = gokord.DB.Where("guild_id = ?", g.ID).Delete(c).Error; err != nil {
					utils.SendAlert(
						"user/level.go - Removing user from guild in database", err.Error(),
						"discord_id", c.DiscordID,
						"guild_id", g.ID,
					)
				}
				continue
			}
			wg.Add(1)
			go func() {
				//do things
				wg.Done()
			}()
		}
		wg.Wait() // finish the entire guild before starting another
		utils.SendDebug("Periodic reduce, guild finished", "guild", g.Name)
		time.Sleep(15 * time.Second) // sleep prevents from spamming the Discord API and the database
	}
	utils.SendDebug("Periodic reduce finished", "len(guilds)", len(dg.State.Guilds))
}
