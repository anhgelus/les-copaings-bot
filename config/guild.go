package config

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"strings"
)

type GuildConfig struct {
	ID               uint   `gorm:"primarykey"`
	GuildID          string `gorm:"not null;unique"`
	XpRoles          []XpRole
	DisabledChannels string
	FallbackChannel  string
	DaysXPRemains    uint `gorm:"default:90"` // 30 * 3 = 90 (three months)
}

type XpRole struct {
	ID            uint `gorm:"primarykey"`
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
			return i, &r
		}
	}
	return 0, nil
}
