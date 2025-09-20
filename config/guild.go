package config

import (
	"strings"

	"github.com/anhgelus/gokord"
	discordgo "github.com/nyttikord/gokord"
)

type GuildConfig struct {
	ID               uint   `gorm:"primarykey"`
	GuildID          string `gorm:"not null;unique"`
	XpRoles          []XpRole
	DisabledChannels string
	FallbackChannel  string
	DaysXPRemains    uint `gorm:"default:90"` // 30 * 3 = 90 (three months)
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

func (cfg *GuildConfig) IsDisabled(s *discordgo.Session, channelID string) bool {
	ok := true
	s.LogInfo("Configuration: %s", cfg.DisabledChannels)
	s.LogInfo("Channel %s, ok %t", channelID, ok)
	for channelID != "" && ok {
		s.LogInfo("Channel %s, ok %t", channelID, ok)
		ok = !strings.Contains(cfg.DisabledChannels, channelID)
		c, err := s.State.Channel(channelID)
		if err != nil {
			s.LogError(err, "Unable to find channel %s in state", c)
			c, err = s.ChannelAPI().Channel(channelID)
			if err == nil {
				s.State.ChannelAdd(c)
			} else {
				s.LogError(err, "Unable to fetch channel %s", s)
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
