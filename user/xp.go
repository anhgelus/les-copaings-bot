package user

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
)

func (c *Copaing) AddXP(s *discordgo.Session, m *discordgo.Member, xp uint, fn func(uint, uint)) {
	pastLevel := exp.Level(c.XP)
	old := c.XP
	c.XP += xp
	if err := c.Save(); err != nil {
		utils.SendAlert(
			"user/xp.go - Saving user",
			err.Error(),
			"exp",
			c.XP,
			"discord_id",
			c.DiscordID,
			"guild_id",
			c.GuildID,
		)
		c.XP = old
		return
	}
	newLevel := exp.Level(c.XP)
	if newLevel > pastLevel {
		fn(c.XP, newLevel)
		onNewLevel(s, m, newLevel)
	}
}
