package user

import (
	"context"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/common"
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
	Copaing() *Copaing
	GetXP() uint
}

func GetCopaing(ctx context.Context, discordID string, guildID string) *CopaingCached {
	state := GetState(ctx)
	cc, err := state.Copaing(guildID, discordID)
	if err != nil {
		c := Copaing{DiscordID: discordID, GuildID: guildID}
		if err := c.load(ctx); err != nil {
			panic(err)
		}
		cc = FromCopaing(&c)
	}
	return cc
}

func (c *Copaing) load(ctx context.Context) error {
	err := common.GetDB(ctx).
		Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).
		Preload("CopaingXPs").
		FirstOrCreate(c).
		Error
	if err != nil {
		return err
	}
	return err
}

func (c *Copaing) Save(ctx context.Context) error {
	return common.GetDB(ctx).Save(c).Error
}

func (c *Copaing) Delete(ctx context.Context) error {
	db := common.GetDB(ctx)
	err := db.
		Where("copaing_id = ? and guild_id = ?", c.ID, c.GuildID).
		Delete(&CopaingXP{}).
		Error
	if err != nil {
		return err
	}
	return db.Where("guild_id = ? AND discord_id = ?", c.GuildID, c.DiscordID).Delete(c).Error
}
