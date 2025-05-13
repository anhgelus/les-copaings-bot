package config

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"gorm.io/gorm"
	"strings"
)

type GuildConfig struct {
	gorm.Model
	GuildID          string `gorm:"not null;unique"`
	XpRoles          []*XpRole
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
	if err := cfg.Load(); err != nil {
		utils.SendAlert("config/guild.go - Loading guild config", err.Error(), "guild_id", guildID)
		return nil
	}
	return &cfg
}

func (cfg *GuildConfig) Load() error {
	return gokord.DB.Where("guild_id = ?", cfg.GuildID).Preload("XpRoles").FirstOrCreate(cfg).Error
}

func (cfg *GuildConfig) Save() error {
	return gokord.DB.Save(cfg).Error
}

func (cfg *GuildConfig) IsDisabled(channelID string) bool {
	return strings.Contains(cfg.DisabledChannels, channelID)
}

func (cfg *GuildConfig) FindXpRole(roleID string) (int, *XpRole) {
	for i, r := range cfg.XpRoles {
		if r.RoleID == roleID {
			return i, r
		}
	}
	return 0, nil
}
