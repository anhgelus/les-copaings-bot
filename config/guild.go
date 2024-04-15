package config

import (
	"github.com/anhgelus/gokord"
	"gorm.io/gorm"
	"strings"
)

type GuildConfig struct {
	gorm.Model
	GuildID          string `gorm:"not null"`
	XpRoles          []XpRole
	DisabledChannels string
	FallbackChannel  string
}

type XpRole struct {
	gorm.Model
	XP            uint
	RoleID        string
	GuildConfigID uint
}

func GetGuildConfig(guildID string) *GuildConfig {
	cfg := GuildConfig{GuildID: guildID}
	return cfg.Load()
}

func (cfg *GuildConfig) Load() *GuildConfig {
	gokord.DB.Where("guild_id = ?", cfg.GuildID).Preload("XpRoles").FirstOrCreate(cfg)
	return cfg
}

func (cfg *GuildConfig) Save() {
	gokord.DB.Save(cfg)
}

func (cfg *GuildConfig) IsDisabled(channelID string) bool {
	return strings.Contains(cfg.DisabledChannels, channelID)
}

func (cfg *GuildConfig) FindXpRole(roleID string) (int, *XpRole) {
	for i, r := range cfg.XpRoles {
		if r.RoleID == roleID {
			return i, &r
		}
	}
	return 0, nil
}
