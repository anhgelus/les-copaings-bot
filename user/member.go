package user

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"gorm.io/gorm"
)

type Copaing struct {
	gorm.Model
	DiscordID string `gorm:"not null"`
	//XP        []CopaingXP
	XP      uint   `gorm:"default:0"`
	GuildID string `gorm:"not null"`
}

type leftCopaing struct {
	ID         uint
	StopDelete chan<- interface{}
}

//type CopaingXP struct {
//	gorm.Model
//	XP        uint `gorm:"default:0"`
//	CopaingID uint
//}

var (
	leftCopaingsMap = map[string]*leftCopaing{}
)

const (
	LastEvent      = "last_event"
	AlreadyRemoved = "already_removed"
)

func GetCopaing(discordID string, guildID string) *Copaing {
	c := Copaing{DiscordID: discordID, GuildID: guildID}
	if err := c.Load(); err != nil {
		utils.SendAlert(
			"exp/member.go - Loading user",
			err.Error(),
			"discord_id",
			discordID,
			"guild_id",
			guildID,
		)
		return nil
	}
	return &c
}

func (c *Copaing) Load() error {
	// check if user left in the past 48 hours
	k := c.GuildID + ":" + c.DiscordID
	l, ok := leftCopaingsMap[k]
	if !ok || l == nil {
		// if not, common first or create
		return gokord.DB.Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).FirstOrCreate(c).Error
	}
	// else, getting last data
	tmp := Copaing{
		Model: gorm.Model{
			ID: c.ID,
		},
		DiscordID: c.DiscordID,
		GuildID:   c.GuildID,
	}
	if err := gokord.DB.Unscoped().Find(&tmp).Error; err != nil {
		// if error, avoid getting old data and use new one
		utils.SendAlert(
			"exp/member.go - Getting user in soft delete", err.Error(),
			"discord_id", c.DiscordID,
			"guild_id", c.DiscordID,
			"last_id", l.ID,
		)
		return gokord.DB.Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).FirstOrCreate(c).Error
	}
	// resetting internal data
	tmp.Model = gorm.Model{}
	l.StopDelete <- true
	leftCopaingsMap[k] = nil
	// creating new data
	err := gokord.DB.Create(&tmp).Error
	if err != nil {
		return err
	}
	// delete old data
	if err = gokord.DB.Unscoped().Delete(&tmp).Error; err != nil {
		utils.SendAlert(
			"exp/member.go - Deleting user in soft delete", err.Error(),
			"discord_id", c.DiscordID,
			"guild_id", c.DiscordID,
			"last_id", l.ID,
		)
	}
	return nil
}

func (c *Copaing) Save() error {
	return gokord.DB.Save(c).Error
}

func (c *Copaing) GenKey(key string) string {
	return fmt.Sprintf("%s:%s:%s", c.GuildID, c.DiscordID, key)
}
