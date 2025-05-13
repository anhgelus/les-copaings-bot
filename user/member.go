package user

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"time"
)

type Copaing struct {
	ID        uint        `gorm:"primarykey"`
	DiscordID string      `gorm:"not null"`
	XP        []CopaingXP `gorm:"constraint:OnDelete:SET NULL;"`
	GuildID   string      `gorm:"not null"`
}

type CopaingXP struct {
	ID        uint   `gorm:"primarykey"`
	XP        uint   `gorm:"default:0"`
	CopaingID uint   `gorm:"not null;constraint:OnDelete:CASCADE;"`
	GuildID   string `gorm:"not null;"`
	CreatedAt time.Time
}

const (
	LastEvent      = "last_event"
	AlreadyRemoved = "already_removed"
)

func GetCopaing(discordID string, guildID string) *Copaing {
	c := Copaing{DiscordID: discordID, GuildID: guildID}
	if err := c.Load(); err != nil {
		utils.SendAlert(
			"user/member.go - Loading user",
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
	return gokord.DB.
		Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).
		Preload("XP").
		FirstOrCreate(c).
		Error
}

func (c *Copaing) GetXP() (uint, error) {
	cfg := config.GetGuildConfig(c.GuildID)
	xp := uint(0)
	y, m, d := time.Unix(time.Now().Unix()-int64(cfg.DaysXPRemains*24*60*60), 0).Date()
	rows, err := gokord.DB.
		Model(&CopaingXP{}).
		Where(fmt.Sprintf("created_at >= '%d-%d-%d' and guild_id = ? and discord_id = ?", y, m, d), c.GuildID, c.DiscordID).
		Rows()
	defer rows.Close()
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var cXP CopaingXP
		err = gokord.DB.ScanRows(rows, &cXP)
		if err != nil {
			utils.SendAlert("user/member.go - Scaning rows", err.Error(), "discord_id", c.DiscordID, "guild_id", c.GuildID)
			continue
		}
		xp += cXP.XP
	}
	return xp, nil
}

func (c *Copaing) Save() error {
	return gokord.DB.Save(c).Error
}

func (c *Copaing) GenKey(key string) string {
	return fmt.Sprintf("%s:%s:%s", c.GuildID, c.DiscordID, key)
}

func (c *Copaing) Delete() error {
	return gokord.DB.Where("guild_id = ? AND discord_id = ?", c.GuildID, c.DiscordID).Delete(c).Error
}
