package user

import (
	"time"

	"github.com/anhgelus/gokord"
)

type Copaing struct {
	ID         uint        `gorm:"primarykey"`
	DiscordID  string      `gorm:"not null"`
	CopaingXPs []CopaingXP `gorm:"constraint:OnDelete:SET NULL;"`
	GuildID    string      `gorm:"not null"`
}

type CopaingXP struct {
	ID        uint `gorm:"primarykey"`
	XP        uint `gorm:"default:0"`
	CopaingID uint
	GuildID   string `gorm:"not null;"`
	CreatedAt time.Time
}

type CopaingAccess interface {
	ToCopaing() *Copaing
	GetXP() uint
}

func GetCopaing(discordID string, guildID string) *Copaing {
	c := Copaing{DiscordID: discordID, GuildID: guildID}
	if err := c.Load(); err != nil {
		panic(err)
	}
	return &c
}

func (c *Copaing) Load() error {
	return gokord.DB.
		Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).
		Preload("CopaingXPs").
		FirstOrCreate(c).
		Error
}

func (c *Copaing) Save() error {
	return gokord.DB.Save(c).Error
}

func (c *Copaing) Delete() error {
	return gokord.DB.Where("guild_id = ? AND discord_id = ?", c.GuildID, c.DiscordID).Delete(c).Error
}
