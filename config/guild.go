package config

import (
	"strings"

	"github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord/bot"
)

type GuildConfig struct {
	ID               uint   `gorm:"primarykey"`
	GuildID          string `gorm:"not null;unique"`
	XpRoles          []XpRole
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

func GetGuildConfig(guildID string) *GuildConfig {
	cfg := GuildConfig{GuildID: guildID}
	if err := cfg.Load(); err != nil {
		panic(err)
	}
	return &cfg
}

func (cfg *GuildConfig) Load() error {
	return gokord.DB.Where("guild_id = ?", cfg.GuildID).Preload("XpRoles").FirstOrCreate(cfg).Error
}

func (cfg *GuildConfig) Save() error {
	return gokord.DB.Save(cfg).Error
}

func (cfg *GuildConfig) IsDisabled(s bot.Session, channelID string) bool {
	ok := true
	for channelID != "" && ok {
		ok = !strings.Contains(cfg.DisabledChannels, channelID)
		c, err := s.ChannelAPI().State.Channel(channelID)
		if err != nil {
			s.Logger().Error("unable to find channel %s in state", "error", err, "channel", c)
			c, err = s.ChannelAPI().Channel(channelID)
			if err == nil {
				err = s.ChannelAPI().State.ChannelAdd(c)
				if err != nil {
					s.Logger().Error("unable to add channel to state", "error", err, "channel", c)
				}
			} else {
				s.Logger().Error("unable to fetch channel", "error", err, "channel", c)
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

func (cfg *GuildConfig) FindXpRole(roleID string) (int, *XpRole) {
	for i, r := range cfg.XpRoles {
		if r.RoleID == roleID {
			return i, &r
		}
	}
	return 0, nil
}

func (cfg *GuildConfig) FindXpRoleID(ID uint) (int, *XpRole) {
	for i, r := range cfg.XpRoles {
		if r.ID == ID {
			return i, &r
		}
	}
	return -1, nil
}
