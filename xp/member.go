package xp

import (
	"github.com/anhgelus/gokord"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

type Copaing struct {
	gorm.Model
	DiscordID string
	XP        uint
	GuildID   string
}

func (c *Copaing) Load() *Copaing {
	gokord.DB.Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).FirstOrCreate(c)
	return c
}

func (c *Copaing) Save() {
	gokord.DB.Save(c)
}

func (c *Copaing) AddXP(s *discordgo.Session, xp uint, fn func(uint, uint)) {
	pastLevel := Level(c.XP)
	c.XP += xp
	c.Save()
	newLevel := Level(c.XP)
	if newLevel > pastLevel {
		fn(c.XP, newLevel)
		onNewLevel(s, newLevel)
	}
}
