package config

import (
	"context"
	"strings"

	"git.anhgelus.world/anhgelus/les-copaings-bot/common"
	"github.com/nyttikord/gokord/bot"
)

type Guild struct {
	ID               uint     `gorm:"primarykey"`
	GuildID          string   `gorm:"not null;unique"`
	XpRoles          []XpRole `gorm:"foreignKey:GuildConfigID"`
	DisabledChannels string
	FallbackChannel  string
	DaysXPRemains    uint `gorm:"default:90"` // 30 * 3 = 90 (three months)
	RrMessages       []RoleReactMessage
}

type RoleReactMessage struct {
	ID            uint   `gorm:"primarykey"`
	MessageID     string `gorm:"not null;unique"`
	ChannelID     string
	GuildID       string
	Note          string
	Roles         []*RoleReact
	GuildConfigID uint
}

type RoleReact struct {
	ID                 uint `gorm:"primarykey"`
	Reaction           string
	RoleID             string
	RoleReactMessageID uint
	CounterID          uint `gorm:"-"`
}

func GetGuildConfig(ctx context.Context, guildID string) *Guild {
	cfg := Guild{GuildID: guildID}
	if err := cfg.Load(ctx); err != nil {
		panic(err)
	}
	return &cfg
}

func (cfg *Guild) Load(ctx context.Context) error {
	return common.GetDB(ctx).Where("guild_id = ?", cfg.GuildID).Preload("XpRoles").FirstOrCreate(cfg).Error
}

func (cfg *Guild) Save(ctx context.Context) error {
	return common.GetDB(ctx).Save(cfg).Error
}

func (cfg *Guild) IsDisabled(ctx context.Context, dg bot.Session, channelID string) bool {
	ok := true
	for channelID != "" && ok {
		ok = !strings.Contains(cfg.DisabledChannels, channelID)
		c, err := dg.ChannelAPI().State.Channel(channelID)
		if err != nil {
			bot.Logger(ctx).Error("unable to find channel %s in state", "error", err, "channel", c)
			c, err = dg.ChannelAPI().Channel(channelID).Do(ctx)
			if err != nil {
				bot.Logger(ctx).Error("unable to fetch channel", "error", err, "channel", c)
				return false
			}
		}
		if err != nil {
			return false
		}
		channelID = c.ParentID
	}
	return !ok
}

func (cfg *Guild) FindXpRole(roleID string) (int, *XpRole) {
	for i, r := range cfg.XpRoles {
		if r.RoleID == roleID {
			return i, &r
		}
	}
	return 0, nil
}

func (cfg *Guild) FindXpRoleID(ID uint) (int, *XpRole) {
	for i, r := range cfg.XpRoles {
		if r.ID == ID {
			return i, &r
		}
	}
	return -1, nil
}
