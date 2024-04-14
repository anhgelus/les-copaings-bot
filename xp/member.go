package xp

import (
	"github.com/anhgelus/gokord"
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
