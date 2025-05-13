package user

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
)

func (c *Copaing) AddXP(s *discordgo.Session, m *discordgo.Member, xp uint, fn func(uint, uint)) {
	old, err := c.GetXP()
	pastLevel := exp.Level(old)
	c.XP = append(c.XP, CopaingXP{CopaingID: c.ID, XP: xp, GuildID: c.GuildID})
	if err = c.Save(); err != nil {
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
		return
	}
	newLevel := exp.Level(old + xp)
	if newLevel > pastLevel {
		fn(old+xp, newLevel)
		onNewLevel(s, m, newLevel)
	}
}
